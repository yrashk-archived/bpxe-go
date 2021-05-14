// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package catch

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

type processEventMessage struct {
	event event.ProcessEvent
}

func (m processEventMessage) message() {}

type Node struct {
	flow_node.T
	element         *bpmn.CatchEvent
	runnerChannel   chan message
	activated       bool
	awaitingActions []chan flow_node.Action
	eventInstances  []event.Instance
	matchedEvents   []bool
}

func New(process *bpmn.Process, definitions *bpmn.Definitions,
	catchEvent *bpmn.CatchEvent, eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource, tracer *tracing.Tracer, flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup, instanceBuilder event.InstanceBuilder) (node *Node, err error) {
	flowNode, err := flow_node.New(process,
		definitions,
		&catchEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	eventDefinitions := catchEvent.EventDefinitions()
	eventInstances := make([]event.Instance, len(eventDefinitions))

	for i, eventDefinition := range eventDefinitions {
		eventInstances[i] = instanceBuilder.NewEventInstance(eventDefinition)
	}

	node = &Node{
		T:               *flowNode,
		element:         catchEvent,
		runnerChannel:   make(chan message, len(flowNode.Incoming)*2+1),
		activated:       false,
		awaitingActions: make([]chan flow_node.Action, 0),
		eventInstances:  eventInstances,
		matchedEvents:   make([]bool, len(eventDefinitions)),
	}
	go node.runner()
	err = node.EventEgress.RegisterProcessEventConsumer(node)
	if err != nil {
		return
	}
	return
}

func (node *Node) runner() {
loop:
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case processEventMessage:
			if node.activated {
				if len(node.eventInstances) == 0 {
					//lint:ignore SA4006 not sure why it's complaining, `ok` is used
					//nolint:staticcheck
					if _, ok := m.event.(event.NoneEvent); ok {
						goto matched
					}
				} else {
					for i, instance := range node.eventInstances {
						if m.event.MatchesEventInstance(instance) {
							node.matchedEvents[i] = true
							goto matched
						}
					}
				}
				continue loop
			matched:
				for _, matched := range node.matchedEvents {
					if !matched && node.element.ParallelMultiple() {
						continue loop
					}
				}
				awaitingActions := node.awaitingActions
				for _, actionChan := range awaitingActions {
					actionChan <- flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
				}
				node.awaitingActions = make([]chan flow_node.Action, 0)
				node.activated = false
			}
		case nextActionMessage:
			if !node.activated {
				node.activated = true
				node.Tracer.Trace(ActiveListeningTrace{Node: node.element})
			}
			node.awaitingActions = append(node.awaitingActions, m.response)
		default:
		}
	}
}

func (node *Node) ConsumeProcessEvent(
	ev event.ProcessEvent,
) (result event.ConsumptionResult, err error) {
	node.runnerChannel <- processEventMessage{event: ev}
	result = event.Consumed
	return
}

func (node *Node) NextAction(flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Node) Incoming(int) {
}

func (node *Node) Element() bpmn.FlowNodeInterface {
	return node.element
}
