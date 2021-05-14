// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package process

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity"
	"bpxe.org/pkg/flow_node/activity/task"
	"bpxe.org/pkg/flow_node/event/catch"
	"bpxe.org/pkg/flow_node/event/end"
	"bpxe.org/pkg/flow_node/event/start"
	"bpxe.org/pkg/flow_node/gateway/event_based"
	"bpxe.org/pkg/flow_node/gateway/exclusive"
	"bpxe.org/pkg/flow_node/gateway/inclusive"
	"bpxe.org/pkg/flow_node/gateway/parallel"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/tracing"
)

type Instance struct {
	process         *Process
	eventConsumers  []event.ProcessEventConsumer
	Tracer          *tracing.Tracer
	flowNodeMapping *flow_node.FlowNodeMapping
	flowWaitGroup   sync.WaitGroup
	complete        sync.RWMutex
	idGenerator     id.Generator
}

// InstanceOption allows to modify configuration of
// an instance in a flexible fashion (as its just a modification
// function)
type InstanceOption func(instance *Instance)

// WithTracer overrides instance's tracer
func WithTracer(tracer *tracing.Tracer) InstanceOption {
	return func(instance *Instance) {
		instance.Tracer = tracer
	}
}

func (instance *Instance) FlowNodeMapping() *flow_node.FlowNodeMapping {
	return instance.flowNodeMapping
}

func NewInstance(process *Process, options ...InstanceOption) (instance *Instance, err error) {
	eventConsumers := make([]event.ProcessEventConsumer, 0)
	tracer := tracing.NewTracer()
	var idGenerator id.Generator
	idGenerator, err = process.GeneratorBuilder.NewIdGenerator(tracer)
	if err != nil {
		return
	}
	instance = &Instance{
		process:         process,
		eventConsumers:  eventConsumers,
		Tracer:          tracer,
		flowNodeMapping: flow_node.NewLockedFlowNodeMapping(),
		idGenerator:     idGenerator,
	}

	// Apply options
	for _, option := range options {
		option(instance)
	}

	for i := range *process.Element.StartEvents() {
		element := &(*process.Element.StartEvents())[i]
		var startEvent *start.Node
		startEvent, err = start.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup,
			idGenerator)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, startEvent)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.EndEvents() {
		element := &(*process.Element.EndEvents())[i]
		var endEvent *end.Node
		endEvent, err = end.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, endEvent)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.IntermediateCatchEvents() {
		element := &(*process.Element.IntermediateCatchEvents())[i]
		var intermediateCatchEvent *catch.Node
		intermediateCatchEvent, err = catch.New(process.Element,
			process.Definitions, &element.CatchEvent, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup,
			process)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, intermediateCatchEvent)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.Tasks() {
		element := &(*process.Element.Tasks())[i]
		var aTask *activity.Harness
		aTask, err = activity.NewHarness(process.Element, process.Definitions,
			&element.FlowNode, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup,
			idGenerator, task.NewTask(element), process,
		)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, aTask)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.ExclusiveGateways() {
		element := &(*process.Element.ExclusiveGateways())[i]
		var exclusiveGateway *exclusive.Node
		exclusiveGateway, err = exclusive.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, exclusiveGateway)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.InclusiveGateways() {
		element := &(*process.Element.InclusiveGateways())[i]
		var inclusiveGateway *inclusive.Node
		inclusiveGateway, err = inclusive.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, inclusiveGateway)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.ParallelGateways() {
		element := &(*process.Element.ParallelGateways())[i]
		var parallelGateway *parallel.Node
		parallelGateway, err = parallel.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, parallelGateway)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.EventBasedGateways() {
		element := &(*process.Element.EventBasedGateways())[i]
		var eventBasedGateway *event_based.Node
		eventBasedGateway, err = event_based.New(process.Element, process.Definitions,
			element, instance, instance, instance.Tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, eventBasedGateway)
		if err != nil {
			return
		}
	}

	instance.flowNodeMapping.Finalize()

	return
}

func (instance *Instance) ConsumeProcessEvent(ev event.ProcessEvent) (result event.ConsumptionResult, err error) {
	result, err = event.ForwardProcessEvent(ev, &instance.eventConsumers)
	return
}

func (instance *Instance) RegisterProcessEventConsumer(ev event.ProcessEventConsumer) (err error) {
	instance.eventConsumers = append(instance.eventConsumers, ev)
	return
}

func (instance *Instance) Run() (err error) {
	lockChan := make(chan bool)
	go func() {
		traces := instance.Tracer.Subscribe()
		instance.complete.Lock()
		lockChan <- true
		/* 13.4.6 End Events:

		The Process instance is [...] completed, if
		and only if the following two conditions
		hold:

		(1) All start nodes of the Process have been
		visited. More precisely, all Start Events
		have been triggered (1.1), and for all
		starting Event-Based Gateways, one of the
		associated Events has been triggered (1.2).

		(2) There is no token remaining within the
		Process instance
		*/
		startEventsActivated := make([]*bpmn.StartEvent, 0)

		// So, at first, we wait for (1.1) to occur
		// [(1.2) will be addded when we actually support them]

		for {
			if len(startEventsActivated) == len(*instance.process.Element.StartEvents()) {
				break
			}

			trace := <-traces

			switch t := trace.(type) {
			case flow.FlowTerminationTrace:
				switch flowNode := t.Source.(type) {
				case *bpmn.StartEvent:
					startEventsActivated = append(startEventsActivated, flowNode)
				default:
				}
			case flow.FlowTrace:
				switch flowNode := t.Source.(type) {
				case *bpmn.StartEvent:
					startEventsActivated = append(startEventsActivated, flowNode)
				default:
				}
			default:
			}
		}

		instance.Tracer.Unsubscribe(traces)

		// Then, we're waiting for (2) to occur
		instance.flowWaitGroup.Wait()
		// Send out a cease flow trace
		instance.Tracer.Trace(flow.CeaseFlowTrace{})
		instance.complete.Unlock()
	}()
	<-lockChan
	close(lockChan)
	evt := event.MakeStartEvent()
	_, err = instance.ConsumeProcessEvent(&evt)
	if err != nil {
		return
	}
	return
}

// Waits until the instance is complete. Returns true if the instance was complete,
// false if the context signalled `Done`
func (instance *Instance) WaitUntilComplete(ctx context.Context) (complete bool) {
	signal := make(chan bool)
	go func() {
		instance.complete.Lock()
		signal <- true
		instance.complete.Unlock()
	}()
	select {
	case <-ctx.Done():
		complete = false
	case <-signal:
		complete = true
	}
	return
}
