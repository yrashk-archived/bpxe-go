// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package start

import (
	"context"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/data"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
}

func (m nextActionMessage) message() {}

type startMessage struct{}

func (m startMessage) message() {}

type eventMessage struct {
	event event.Event
}

func (m eventMessage) message() {}

type Node struct {
	*flow_node.Wiring
	element          *bpmn.StartEvent
	runnerChannel    chan message
	activated        bool
	idGenerator      id.Generator
	itemAwareLocator data.ItemAwareLocator
	eventInstances   []event.Instance
}

func New(ctx context.Context, wiring *flow_node.Wiring, startEvent *bpmn.StartEvent,
	idGenerator id.Generator, itemAwareLocator data.ItemAwareLocator,
) (node *Node, err error) {
	eventDefinitions := startEvent.EventDefinitions()
	eventInstances := make([]event.Instance, len(eventDefinitions))

	for i, eventDefinition := range eventDefinitions {
		eventInstances[i] = wiring.EventInstanceBuilder.NewEventInstance(eventDefinition)
	}

	node = &Node{
		Wiring:           wiring,
		element:          startEvent,
		runnerChannel:    make(chan message, len(wiring.Incoming)*2+1),
		activated:        false,
		idGenerator:      idGenerator,
		itemAwareLocator: itemAwareLocator,
		eventInstances:   eventInstances,
	}
	sender := node.Tracer.RegisterSender()
	go node.runner(ctx, sender)
	err = node.EventEgress.RegisterEventConsumer(node)
	if err != nil {
		return
	}
	return
}

func (node *Node) runner(ctx context.Context, sender tracing.SenderHandle) {
	defer sender.Done()

	for {
		select {
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case nextActionMessage:
				if !node.activated {
					node.activated = true
					m.response <- flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
				} else {
					m.response <- flow_node.CompleteAction{}
				}
			case startMessage:
				node.flow(ctx)
			case eventMessage:
				if !node.activated {
					for i := range node.eventInstances {
						if m.event.MatchesEventInstance(node.eventInstances[i]) {
							if !node.element.ParallelMultiple() {
								node.flow(ctx)
							} else {
								node.eventInstances[i] = node.eventInstances[len(node.eventInstances)-1]
								node.eventInstances = node.eventInstances[:len(node.eventInstances)-1]
								if len(node.eventInstances) == 0 {
									node.flow(ctx)
								}
							}
							break
						}
					}
				}
			default:
			}
		case <-ctx.Done():
			node.Tracer.Trace(flow_node.CancellationTrace{Node: node.element})
			return
		}
	}
}

func (node *Node) flow(ctx context.Context) {
	newFlow := flow.New(node.Wiring.Definitions, node, node.Wiring.Tracer,
		node.Wiring.FlowNodeMapping, node.Wiring.FlowWaitGroup, node.idGenerator, nil,
		node.itemAwareLocator,
	)
	newFlow.Start(ctx)
}

func (node *Node) ConsumeEvent(
	ev event.Event,
) (result event.ConsumptionResult, err error) {
	node.runnerChannel <- eventMessage{event: ev}
	result = event.Consumed
	return
}

func (node *Node) Trigger() {
	node.runnerChannel <- startMessage{}
}

func (node *Node) NextAction(flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Node) Element() bpmn.FlowNodeInterface {
	return node.element
}
