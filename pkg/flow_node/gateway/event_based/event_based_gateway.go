// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package event_based

import (
	"context"
	"fmt"
	"sync/atomic"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
	flow     flow_interface.T
}

func (m nextActionMessage) message() {}

type Node struct {
	*flow_node.Wiring
	element       *bpmn.EventBasedGateway
	runnerChannel chan message
	activated     bool
}

func New(ctx context.Context, wiring *flow_node.Wiring, eventBasedGateway *bpmn.EventBasedGateway) (node *Node, err error) {
	node = &Node{
		Wiring:        wiring,
		element:       eventBasedGateway,
		runnerChannel: make(chan message, len(wiring.Incoming)*2+1),
		activated:     false,
	}
	sender := node.Tracer.RegisterSender()
	go node.runner(ctx, sender)
	return
}

func (node *Node) runner(ctx context.Context, sender tracing.SenderHandle) {
	defer sender.Done()

	for {
		select {
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case nextActionMessage:
				var first int32 = 0
				sequenceFlows := flow_node.AllSequenceFlows(&node.Outgoing)
				terminationChannels := make(map[bpmn.IdRef]chan bool)
				for _, sequenceFlow := range sequenceFlows {
					if idPtr, present := sequenceFlow.Id(); present {
						terminationChannels[*idPtr] = make(chan bool)
					} else {
						node.Tracer.Trace(tracing.ErrorTrace{Error: errors.NotFoundError{
							Expected: fmt.Sprintf("id for %#v", sequenceFlow),
						}})
					}
				}
				m.response <- flow_node.FlowAction{
					Terminate: func(sequenceFlowId *bpmn.IdRef) chan bool {
						return terminationChannels[*sequenceFlowId]
					},
					SequenceFlows: sequenceFlows,
					ActionTransformer: func(sequenceFlowId *bpmn.IdRef, action flow_node.Action) flow_node.Action {
						// only first one is to flow
						if atomic.CompareAndSwapInt32(&first, 0, 1) {
							node.Tracer.Trace(DeterminationMadeTrace{Element: node.element})
							for terminationCandidateId, ch := range terminationChannels {
								if sequenceFlowId != nil && terminationCandidateId != *sequenceFlowId {
									ch <- true
								}
								close(ch)
							}
							terminationChannels = make(map[bpmn.IdRef]chan bool)
							return action
						} else {
							return flow_node.CompleteAction{}
						}
					},
				}
			default:
			}
		case <-ctx.Done():
			node.Tracer.Trace(flow_node.CancellationTrace{Node: node.element})
			return
		}
	}
}

func (node *Node) NextAction(flow flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flow: flow}
	return response
}

func (node *Node) Element() bpmn.FlowNodeInterface {
	return node.element
}
