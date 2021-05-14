// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package exclusive_gateway

import (
	"fmt"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

type NoEffectiveSequenceFlows struct {
	*bpmn.ExclusiveGateway
}

func (e NoEffectiveSequenceFlows) Error() string {
	ownId := "<unnamed>"
	if ownIdPtr, present := e.ExclusiveGateway.Id(); present {
		ownId = *ownIdPtr
	}
	return fmt.Sprintf("No effective sequence flows found in exclusive gateway `%v`", ownId)
}

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
	flow     flow_interface.T
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type probingReport struct {
	result []int
	flowId id.Id
}

func (m probingReport) message() {}

type ExclusiveGateway struct {
	flow_node.FlowNode
	element                 *bpmn.ExclusiveGateway
	runnerChannel           chan message
	defaultSequenceFlow     *sequence_flow.SequenceFlow
	nonDefaultSequenceFlows []*sequence_flow.SequenceFlow
	probing                 map[id.Id]*chan flow_node.Action
}

func NewExclusiveGateway(process *bpmn.Process,
	definitions *bpmn.Definitions,
	exclusiveGateway *bpmn.ExclusiveGateway,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *ExclusiveGateway, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&exclusiveGateway.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}

	var defaultSequenceFlow *sequence_flow.SequenceFlow

	if seqFlow, present := exclusiveGateway.Default(); present {
		if node, found := flowNode.Process.FindBy(bpmn.ExactId(*seqFlow).
			And(bpmn.ElementType((*bpmn.SequenceFlow)(nil)))); found {
			defaultSequenceFlow = new(sequence_flow.SequenceFlow)
			*defaultSequenceFlow = sequence_flow.MakeSequenceFlow(
				node.(*bpmn.SequenceFlow),
				definitions,
			)
		} else {
			err = errors.NotFoundError{
				Expected: fmt.Sprintf("default sequence flow with ID %s", *seqFlow),
			}
			return nil, err
		}
	}

	nonDefaultSequenceFlows := flow_node.AllSequenceFlows(&flowNode.Outgoing,
		func(sequenceFlow *sequence_flow.SequenceFlow) bool {
			if defaultSequenceFlow == nil {
				return false
			}
			return *sequenceFlow == *defaultSequenceFlow
		},
	)

	node = &ExclusiveGateway{
		FlowNode:                *flowNode,
		element:                 exclusiveGateway,
		runnerChannel:           make(chan message, len(flowNode.Incoming)*2+1),
		nonDefaultSequenceFlows: nonDefaultSequenceFlows,
		defaultSequenceFlow:     defaultSequenceFlow,
		probing:                 make(map[id.Id]*chan flow_node.Action),
	}
	go node.runner()
	return
}

func (node *ExclusiveGateway) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case probingReport:
			if response, ok := node.probing[m.flowId]; ok {
				if response == nil {
					// Reschedule, there's no next action yet
					go func() {
						node.runnerChannel <- m
					}()
					continue
				}
				delete(node.probing, m.flowId)
				flow := make([]*sequence_flow.SequenceFlow, 0)
				for _, i := range m.result {
					flow = append(flow, node.nonDefaultSequenceFlows[i])
					break
				}
				switch len(flow) {
				case 0:
					// no successful non-default sequence flows
					if node.defaultSequenceFlow == nil {
						// exception (Table 13.2)
						node.FlowNode.Tracer.Trace(tracing.ErrorTrace{
							Error: NoEffectiveSequenceFlows{
								ExclusiveGateway: node.element,
							},
						})
					} else {
						// default
						*response <- flow_node.FlowAction{
							SequenceFlows:      []*sequence_flow.SequenceFlow{node.defaultSequenceFlow},
							UnconditionalFlows: []int{0},
						}
					}
				case 1:
					*response <- flow_node.FlowAction{
						SequenceFlows:      flow,
						UnconditionalFlows: []int{0},
					}
				default:
					node.FlowNode.Tracer.Trace(tracing.ErrorTrace{
						Error: errors.InvalidArgumentError{
							Expected: fmt.Sprintf("maximum 1 outgoing exclusive gateway (%s) flow",
								node.FlowNode.Id),
							Actual: len(flow),
						},
					})
				}
			} else {
				node.FlowNode.Tracer.Trace(tracing.ErrorTrace{
					Error: errors.InvalidStateError{
						Expected: fmt.Sprintf("probing[%s] is to be present (exclusive gateway %s)",
							m.flowId.String(), node.FlowNode.Id),
					},
				})
			}
		case nextActionMessage:
			if _, ok := node.probing[m.flow.Id()]; ok {
				node.probing[m.flow.Id()] = &m.response
				// and now we wait until the probe has returned
			} else {
				node.probing[m.flow.Id()] = nil
				m.response <- flow_node.ProbeAction{
					SequenceFlows: node.nonDefaultSequenceFlows,
					ProbeReport: func(indices []int) {
						node.runnerChannel <- probingReport{
							result: indices,
							flowId: m.flow.Id(),
						}
					},
				}
			}
		default:
		}
	}
}

func (node *ExclusiveGateway) NextAction(flow flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flow: flow}
	return response
}

func (node *ExclusiveGateway) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *ExclusiveGateway) Element() bpmn.FlowNodeInterface {
	return node.element
}
