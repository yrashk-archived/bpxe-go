// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package activity

import (
	"context"
	"sync"
	"sync/atomic"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/data"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/event/catch"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	flow     flow_interface.T
	response chan chan flow_node.Action
}

func (m nextActionMessage) message() {}

type Harness struct {
	*flow_node.Wiring
	element        bpmn.FlowNodeInterface
	runnerChannel  chan message
	activity       Activity
	activeBoundary <-chan bool
	active         int32
	cancellation   sync.Once
	lock           sync.RWMutex
	eventConsumers []event.ProcessEventConsumer
}

func (node *Harness) ConsumeProcessEvent(ev event.ProcessEvent) (result event.ConsumptionResult, err error) {
	node.lock.RLock()
	defer node.lock.RUnlock()
	if atomic.LoadInt32(&node.active) == 1 {
		result, err = event.ForwardProcessEvent(ev, &node.eventConsumers)
	} else {
		result = event.Consumed
	}
	return
}

func (node *Harness) RegisterProcessEventConsumer(consumer event.ProcessEventConsumer) (err error) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.eventConsumers = append(node.eventConsumers, consumer)
	return
}

func (node *Harness) Activity() Activity {
	return node.activity
}

type Constructor = func(*flow_node.Wiring) (node Activity, err error)

func NewHarness(ctx context.Context,
	wiring *flow_node.Wiring,
	element *bpmn.FlowNode,
	idGenerator id.Generator,
	constructor Constructor,
	instanceBuilder event.InstanceBuilder,
	itemAwareLocator data.ItemAwareLocator,
) (node *Harness, err error) {
	var activity Activity
	activity, err = constructor(wiring)
	if err != nil {
		return
	}

	boundaryEvents := make([]*bpmn.BoundaryEvent, 0)

	for i := range *wiring.Process.BoundaryEvents() {
		boundaryEvent := &(*wiring.Process.BoundaryEvents())[i]
		if *boundaryEvent.AttachedToRef() == wiring.Id {
			boundaryEvents = append(boundaryEvents, boundaryEvent)
		}
	}

	node = &Harness{
		Wiring:         wiring,
		element:        element,
		runnerChannel:  make(chan message, len(wiring.Incoming)*2+1),
		activity:       activity,
		activeBoundary: activity.ActiveBoundary(),
	}

	err = node.EventEgress.RegisterProcessEventConsumer(node)
	if err != nil {
		return
	}

	for i := range boundaryEvents {
		boundaryEvent := boundaryEvents[i]
		var catchEventFlowNode *flow_node.Wiring
		catchEventFlowNode, err = wiring.CloneFor(&boundaryEvent.FlowNode)
		if err != nil {
			return
		}
		// this node becomes event egress
		catchEventFlowNode.EventEgress = node

		var catchEvent *catch.Node
		catchEvent, err = catch.New(ctx, catchEventFlowNode, &boundaryEvent.CatchEvent, instanceBuilder)
		if err != nil {
			return
		} else {
			var actionTransformer flow_node.ActionTransformer
			if boundaryEvent.CancelActivity() {
				actionTransformer = func(sequenceFlowId *bpmn.IdRef, action flow_node.Action) flow_node.Action {
					node.cancellation.Do(func() {
						<-node.activity.Cancel()
					})
					return action
				}
			}
			newFlow := flow.New(node.Definitions, catchEvent, node.Tracer,
				node.FlowNodeMapping, node.FlowWaitGroup, idGenerator, actionTransformer, itemAwareLocator)
			newFlow.Start(ctx)
		}
	}
	sender := node.Tracer.RegisterSender()
	go node.runner(ctx, sender)
	return
}

func (node *Harness) runner(ctx context.Context, sender tracing.SenderHandle) {
	defer sender.Done()

	for {
		select {
		case activeBoundary := <-node.activeBoundary:
			if activeBoundary && atomic.LoadInt32(&node.active) == 0 {
				// Opening active boundary
				atomic.StoreInt32(&node.active, 1)
				node.Tracer.Trace(ActiveBoundaryTrace{Start: activeBoundary, Node: node.activity.Element()})
			} else if !activeBoundary && atomic.LoadInt32(&node.active) == 1 {
				// Closing active boundary
				atomic.StoreInt32(&node.active, 0)
				node.Tracer.Trace(ActiveBoundaryTrace{Start: activeBoundary, Node: node.activity.Element()})
			}
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case nextActionMessage:
				m.response <- node.activity.NextAction(m.flow)
			default:
			}
		case <-ctx.Done():
			node.Tracer.Trace(flow_node.CancellationTrace{Node: node.element})
			return
		}
	}
}

func (node *Harness) NextAction(flow flow_interface.T) chan flow_node.Action {
	response := make(chan chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{flow: flow, response: response}
	return <-response
}

func (node *Harness) Element() bpmn.FlowNodeInterface {
	return node.element
}
