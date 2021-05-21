// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package instance

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
	process                        *bpmn.Process
	Tracer                         *tracing.Tracer
	flowNodeMapping                *flow_node.FlowNodeMapping
	flowWaitGroup                  sync.WaitGroup
	complete                       sync.RWMutex
	idGenerator                    id.Generator
	dataObjectsByName              map[string]data.ItemAware
	dataObjects                    map[bpmn.Id]data.ItemAware
	dataObjectReferencesByName     map[string]data.ItemAware
	dataObjectReferences           map[bpmn.Id]data.ItemAware
	propertiesByName               map[string]data.ItemAware
	properties                     map[bpmn.Id]data.ItemAware
	EventIngress                   event.Consumer
	EventEgress                    event.Source
	idGeneratorBuilder             id.GeneratorBuilder
	eventDefinitionInstanceBuilder event.DefinitionInstanceBuilder
	eventConsumersLock             sync.RWMutex
	eventConsumers                 []event.Consumer
}

func (instance *Instance) ConsumeEvent(ev event.Event) (result event.ConsumptionResult, err error) {
	instance.eventConsumersLock.RLock()
	// We're copying the list of consumers here to ensure that
	// new consumers can subscribe during event forwarding
	eventConsumers := instance.eventConsumers
	instance.eventConsumersLock.RUnlock()
	result, err = event.ForwardEvent(ev, &eventConsumers)
	return
}

func (instance *Instance) RegisterEventConsumer(ev event.Consumer) (err error) {
	instance.eventConsumersLock.Lock()
	instance.eventConsumers = append(instance.eventConsumers, ev)
	instance.eventConsumersLock.Unlock()
	return
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

// Option allows to modify configuration of
// an instance in a flexible fashion (as its just a modification
// function)
//
// It also allows to augment or replace the context.
type Option func(ctx context.Context, instance *Instance) context.Context

// WithTracer overrides instance's tracer
func WithTracer(tracer *tracing.Tracer) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.Tracer = tracer
		return ctx
	}
}

// WithContext will pass a given context to a new instance
// instead of implicitly generated one
func WithContext(newCtx context.Context) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		return newCtx
	}
}

func WithIdGenerator(builder id.GeneratorBuilder) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.idGeneratorBuilder = builder
		return ctx
	}
}

func WithEventIngress(consumer event.Consumer) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.EventIngress = consumer
		return ctx
	}
}

func WithEventEgress(source event.Source) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.EventEgress = source
		return ctx
	}
}

func WitheventDefinitionInstanceBuilder(builder event.DefinitionInstanceBuilder) Option {
	return func(ctx context.Context, instance *Instance) context.Context {
		instance.eventDefinitionInstanceBuilder = builder
		return ctx
	}
}

func (instance *Instance) FlowNodeMapping() *flow_node.FlowNodeMapping {
	return instance.flowNodeMapping
}

func NewInstance(element *bpmn.Process, definitions *bpmn.Definitions, options ...Option) (instance *Instance, err error) {
	instance = &Instance{
		process:                    element,
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

	if instance.idGeneratorBuilder == nil {
		instance.idGeneratorBuilder = id.DefaultIdGeneratorBuilder
	}

	var idGenerator id.Generator
	idGenerator, err = instance.idGeneratorBuilder.NewIdGenerator(ctx, instance.Tracer)
	if err != nil {
		return
	}

	instance.idGenerator = idGenerator

	err = instance.EventEgress.RegisterEventConsumer(instance)
	if err != nil {
		return
	}

	// Item aware elements

	for i := range *instance.process.DataObjects() {
		dataObject := &(*instance.process.DataObjects())[i]
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

	for i := range *instance.process.DataObjectReferences() {
		dataObjectReference := &(*instance.process.DataObjectReferences())[i]
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

	for i := range *instance.process.Properties() {
		property := &(*instance.process.Properties())[i]
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
		return flow_node.New(instance.process,
			definitions,
			element,
			// Event ingress/egress orchestration:
			//
			// Flow nodes will send their message to `instance.EventIngress`
			// (which is typically the model), but consume their messages from
			// `instance`, which is turn a subscriber of `instance.EventEgress`
			// (again, typically, the model).
			//
			// This allows us to use ConsumeEvent on this instance to send
			// events only to the instance (useful for things like event-based
			// process instantiation)
			instance.EventIngress, instance,
			instance.Tracer, instance.flowNodeMapping,
			&instance.flowWaitGroup, instance.eventDefinitionInstanceBuilder)
	}

	var wiring *flow_node.Wiring

	for i := range *instance.process.StartEvents() {
		element := &(*instance.process.StartEvents())[i]
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

	for i := range *instance.process.EndEvents() {
		element := &(*instance.process.EndEvents())[i]
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

	for i := range *instance.process.IntermediateCatchEvents() {
		element := &(*instance.process.IntermediateCatchEvents())[i]
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var intermediateCatchEvent *catch.Node
		intermediateCatchEvent, err = catch.New(ctx, wiring, &element.CatchEvent)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, intermediateCatchEvent)
		if err != nil {
			return
		}
	}

	for i := range *instance.process.Tasks() {
		element := &(*instance.process.Tasks())[i]
		wiring, err = wiringMaker(&element.FlowNode)
		if err != nil {
			return
		}
		var aTask *activity.Harness
		aTask, err = activity.NewHarness(ctx, wiring, &element.FlowNode,
			idGenerator, task.NewTask(ctx, element), instance,
		)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, aTask)
		if err != nil {
			return
		}
	}

	for i := range *instance.process.ExclusiveGateways() {
		element := &(*instance.process.ExclusiveGateways())[i]
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

	for i := range *instance.process.InclusiveGateways() {
		element := &(*instance.process.InclusiveGateways())[i]
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

	for i := range *instance.process.ParallelGateways() {
		element := &(*instance.process.ParallelGateways())[i]
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

	for i := range *instance.process.EventBasedGateways() {
		element := &(*instance.process.EventBasedGateways())[i]
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

// StartWith explicitly starts the instance by triggering a given start event
func (instance *Instance) StartWith(ctx context.Context, startEvent bpmn.StartEventInterface) (err error) {
	flowNode, found := instance.flowNodeMapping.ResolveElementToFlowNode(startEvent)
	elementId := "<unnamed>"
	if idPtr, present := startEvent.Id(); present {
		elementId = *idPtr
	}
	processId := "<unnamed>"
	if idPtr, present := instance.process.Id(); present {
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
	for i := range *instance.process.StartEvents() {
		err = instance.StartWith(ctx, &(*instance.process.StartEvents())[i])
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
			if len(startEventsActivated) == len(*instance.process.StartEvents()) {
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
