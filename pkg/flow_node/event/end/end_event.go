// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package end

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type Node struct {
	flow_node.FlowNode
	element              *bpmn.EndEvent
	activated            bool
	completed            bool
	eventConsumer        event.ProcessEventConsumer
	runnerChannel        chan message
	startEventsActivated []*bpmn.StartEvent
}

func New(process *bpmn.Process,
	definitions *bpmn.Definitions,
	endEvent *bpmn.EndEvent,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *Node, err error) {
	flowNodePtr, err := flow_node.NewFlowNode(
		process,
		definitions,
		&endEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	flowNode := *flowNodePtr
	node = &Node{
		FlowNode:             flowNode,
		element:              endEvent,
		activated:            false,
		completed:            false,
		eventConsumer:        eventIngress,
		runnerChannel:        make(chan message, len(flowNode.Incoming)*2+1),
		startEventsActivated: make([]*bpmn.StartEvent, 0),
	}
	go node.runner()
	return
}

func (node *Node) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case incomingMessage:
			node.activated = true
		case nextActionMessage:
			// If the node hasn't been activated, it's too early
			if !node.activated {
				m.response <- flow_node.NoAction{}
				continue
			}
			// If the node already completed, then we essentially fuse it
			if node.completed {
				m.response <- flow_node.CompleteAction{}
				continue
			}

			if _, err := node.FlowNode.EventIngress.ConsumeProcessEvent(
				event.MakeEndEvent(node.element),
			); err == nil {
				node.completed = true
				m.response <- flow_node.CompleteAction{}
			} else {
				node.FlowNode.Tracer.Trace(tracing.ErrorTrace{Error: err})
			}
		default:
		}
	}
}

func (node *Node) NextAction(flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Node) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *Node) Element() bpmn.FlowNodeInterface {
	return node.element
}
