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
	"fmt"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/data"
	"bpxe.org/pkg/errors"
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
	process                    *Process
	eventConsumers             []event.ProcessEventConsumer
	Tracer                     *tracing.Tracer
	flowNodeMapping            *flow_node.FlowNodeMapping
	flowWaitGroup              sync.WaitGroup
	complete                   sync.RWMutex
	idGenerator                id.Generator
	dataObjectsByName          map[string]data.ItemAware
	dataObjects                map[bpmn.Id]data.ItemAware
	dataObjectReferencesByName map[string]data.ItemAware
	dataObjectReferences       map[bpmn.Id]data.ItemAware
	propertiesByName           map[string]data.ItemAware
	properties                 map[bpmn.Id]data.ItemAware
}

func (instance *Instance) FindItemAwareById(id bpmn.IdRef) (itemAware data.ItemAware, found bool) {
	for k := range instance.dataObjects {
		if k == id {
			found = true
			itemAware = instance.dataObjects[k]
			goto ready
		}
	}
	for k := range instance.dataObjectReferences {
		if k == id {
			found = true
			itemAware = instance.dataObjectReferences[k]
			goto ready
		}
	}
	for k := range instance.properties {
		if k == id {
			found = true
			itemAware = instance.properties[k]
			goto ready
		}
	}
ready:
	return
}

func (instance *Instance) FindItemAwareByName(name string) (itemAware data.ItemAware, found bool) {
	for k := range instance.dataObjectsByName {
		if k == name {
			found = true
			itemAware = instance.dataObjectsByName[k]
			goto ready
		}
	}
	for k := range instance.dataObjectReferencesByName {
		if k == name {
			found = true
			itemAware = instance.dataObjectReferencesByName[k]
			goto ready
		}
	}
	for k := range instance.propertiesByName {
		if k == name {
			found = true
			itemAware = instance.propertiesByName[k]
			goto ready
		}
	}
ready:
	return
}

// InstanceOption allows to modify configuration of
// an instance in a flexible fashion (as its just a modification
// function)
//
// It also allows to augment or replace the context.
type InstanceOption func(ctx context.Context, instance *Instance) context.Context

// WithTracer overrides instance's tracer
func WithTracer(tracer *tracing.Tracer) InstanceOption {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.Tracer = tracer
		return ctx
	}
}

// WithContext will pass a given context to a new instance
// instead of implicitly generated one
func WithContext(newCtx context.Context) InstanceOption {
	return func(ctx context.Context, instance *Instance) context.Context {
		return newCtx
	}
}

func (instance *Instance) FlowNodeMapping() *flow_node.FlowNodeMapping {
	return instance.flowNodeMapping
}

func NewInstance(process *Process, options ...InstanceOption) (instance *Instance, err error) {
	eventConsumers := make([]event.ProcessEventConsumer, 0)

	instance = &Instance{
		process:                    process,
		eventConsumers:             eventConsumers,
		flowNodeMapping:            flow_node.NewLockedFlowNodeMapping(),
		dataObjectsByName:          make(map[string]data.ItemAware),
		dataObjectReferencesByName: make(map[string]data.ItemAware),
		propertiesByName:           make(map[string]data.ItemAware),
		dataObjects:                make(map[string]data.ItemAware),
		dataObjectReferences:       make(map[string]data.ItemAware),
		properties:                 make(map[string]data.ItemAware),
	}

	ctx := context.Background()

	// Apply options
	for _, option := range options {
		ctx = option(ctx, instance)
	}

	if instance.Tracer == nil {
		instance.Tracer = tracing.NewTracer(ctx)
	}

	var idGenerator id.Generator
	idGenerator, err = process.GeneratorBuilder.NewIdGenerator(ctx, instance.Tracer)
	if err != nil {
		return
	}

	instance.idGenerator = idGenerator

	// Item aware elements

	for i := range *process.Element.DataObjects() {
		dataObject := &(*process.Element.DataObjects())[i]
		var name string
		if namePtr, present := dataObject.Name(); present {
			name = *namePtr
		} else {
			name = idGenerator.New().String()
		}
		container := data.NewContainer(ctx, dataObject)
		instance.dataObjectsByName[name] = container
		if idPtr, present := dataObject.Id(); present {
			instance.dataObjects[*idPtr] = container
		}
	}

	for i := range *process.Element.DataObjectReferences() {
		dataObjectReference := &(*process.Element.DataObjectReferences())[i]
		var name string
		if namePtr, present := dataObjectReference.Name(); present {
			name = *namePtr
		} else {
			name = idGenerator.New().String()
		}
		var container data.ItemAware
		if dataObjPtr, present := dataObjectReference.DataObjectRef(); present {
			for dataObjectId := range instance.dataObjects {
				if dataObjectId == *dataObjPtr {
					container = instance.dataObjects[dataObjectId]
					break
				}
			}
			if container == nil {
				err = errors.NotFoundError{
					Expected: fmt.Sprintf("data object with ID %s", *dataObjPtr),
				}
				return
			}
		} else {
			err = errors.InvalidArgumentError{
				Expected: "data object reference to have dataObjectRef",
				Actual:   dataObjectReference,
			}
			return
		}
		instance.dataObjectReferencesByName[name] = container
		if idPtr, present := dataObjectReference.Id(); present {
			instance.dataObjectReferences[*idPtr] = container
		}
	}

	for i := range *process.Element.Properties() {
		property := &(*process.Element.Properties())[i]
		var name string
		if namePtr, present := property.Name(); present {
			name = *namePtr
		} else {
			name = idGenerator.New().String()
		}
		container := data.NewContainer(ctx, property)
		instance.propertiesByName[name] = container
		if idPtr, present := property.Id(); present {
			instance.properties[*idPtr] = container
		}
	}

	// Flow nodes

	wiringMaker := func(element *bpmn.FlowNode) (*flow_node.Wiring, error) {
		return flow_node.New(process.Element,
			process.Definitions,
			element, instance, instance,
			instance.Tracer, instance.flowNodeMapping,
			&instance.flowWaitGroup)
	}

	var wiring *flow_node.Wiring

	for i := range *process.Element.StartEvents() {
		element := &(*process.Element.StartEvents())[i]
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var startEvent *start.Node
		startEvent, err = start.New(ctx, wiring, element, idGenerator, instance)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var endEvent *end.Node
		endEvent, err = end.New(ctx, wiring, element)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var intermediateCatchEvent *catch.Node
		intermediateCatchEvent, err = catch.New(ctx, wiring, &element.CatchEvent, process)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var aTask *activity.Harness
		aTask, err = activity.NewHarness(ctx, wiring, &element.FlowNode,
			idGenerator, task.NewTask(ctx, element), process, instance,
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var exclusiveGateway *exclusive.Node
		exclusiveGateway, err = exclusive.New(ctx, wiring, element)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var inclusiveGateway *inclusive.Node
		inclusiveGateway, err = inclusive.New(ctx, wiring, element)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		parallelGateway, err = parallel.New(ctx, wiring, element)
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
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var eventBasedGateway *event_based.Node
		eventBasedGateway, err = event_based.New(ctx, wiring, element)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, eventBasedGateway)
		if err != nil {
			return
		}
	}

	instance.flowNodeMapping.Finalize()

	// StartAll cease flow monitor
	sender := instance.Tracer.RegisterSender()
	go instance.ceaseFlowMonitor()(ctx, sender)

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

// StartWith explicitly starts the instance by triggering a given start event
func (instance *Instance) StartWith(ctx context.Context, startEvent bpmn.StartEventInterface) (err error) {
	flowNode, found := instance.flowNodeMapping.ResolveElementToFlowNode(startEvent)
	elementId := "<unnamed>"
	if idPtr, present := startEvent.Id(); present {
		elementId = *idPtr
	}
	processId := "<unnamed>"
	if idPtr, present := instance.process.Element.Id(); present {
		processId = *idPtr
	}
	if !found {
		err = errors.NotFoundError{Expected: fmt.Sprintf("start event %s in process %s", elementId, processId)}
		return
	}
	startEventNode, ok := flowNode.(*start.Node)
	if !ok {
		err = errors.RequirementExpectationError{
			Expected: fmt.Sprintf("start event %s flow node in process %s to be of type start.Node", elementId, processId),
			Actual:   fmt.Sprintf("%T", flowNode),
		}
		return
	}
	startEventNode.Trigger()
	return
}

// StartAll explicitly starts the instance by triggering all start events, if any
func (instance *Instance) StartAll(ctx context.Context) (err error) {
	for i := range *instance.process.Element.StartEvents() {
		err = instance.StartWith(ctx, &(*instance.process.Element.StartEvents())[i])
		if err != nil {
			return
		}
	}
	return
}

func (instance *Instance) ceaseFlowMonitor() func(ctx context.Context, sender tracing.SenderHandle) {
	// Subscribing to traces early as otherwise events produced
	// after the goroutine below is started are not going to be
	// sent to it.
	traces := instance.Tracer.Subscribe()
	instance.complete.Lock()
	return func(ctx context.Context, sender tracing.SenderHandle) {
		defer sender.Done()
		defer instance.complete.Unlock()

		/* 13.4.6 End Events:

		The Process instance is [...] completed, if
		and only if the following two conditions
		hold:

		(1) All start nodes of the Process have been
		visited. More precisely, all StartAll Events
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

			select {
			case trace := <-traces:
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
			case <-ctx.Done():
				instance.Tracer.Unsubscribe(traces)
				return
			}
		}

		instance.Tracer.Unsubscribe(traces)

		// Then, we're waiting for (2) to occur
		waitIsOver := make(chan struct{})
		go func() {
			instance.flowWaitGroup.Wait()
			close(waitIsOver)
		}()
		select {
		case <-waitIsOver:
			// Send out a cease flow trace
			instance.Tracer.Trace(flow.CeaseFlowTrace{})
		case <-ctx.Done():
		}

	}
}

// WaitUntilComplete waits until the instance is complete.
// Returns true if the instance was complete, false if the context signalled `Done`
func (instance *Instance) WaitUntilComplete(ctx context.Context) (complete bool) {
	signal := make(chan bool)
	go func() {
		instance.complete.Lock()
		defer instance.complete.Unlock()
		signal <- true
	}()
	select {
	case <-ctx.Done():
		complete = false
	case <-signal:
		complete = true
	}
	return
}
