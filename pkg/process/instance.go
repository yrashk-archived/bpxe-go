package process

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity/task"
	"bpxe.org/pkg/flow_node/event/end_event"
	"bpxe.org/pkg/flow_node/event/intermediate_catch_event"
	"bpxe.org/pkg/flow_node/event/start_event"
	"bpxe.org/pkg/flow_node/gateway/event_based_gateway"
	"bpxe.org/pkg/flow_node/gateway/exclusive_gateway"
	"bpxe.org/pkg/flow_node/gateway/parallel_gateway"
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
	idGenerator     id.IdGenerator
}

func NewInstance(process *Process) (instance *Instance, err error) {
	eventConsumers := make([]event.ProcessEventConsumer, 0)
	tracer := tracing.NewTracer()
	var idGenerator id.IdGenerator
	idGenerator, err = process.IdGeneratorBuilder.NewIdGenerator(tracer)
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

	for i := range *process.Element.StartEvents() {
		element := &(*process.Element.StartEvents())[i]
		var startEvent *start_event.StartEvent
		startEvent, err = start_event.NewStartEvent(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup,
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
		var endEvent *end_event.EndEvent
		endEvent, err = end_event.NewEndEvent(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
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
		var intermediateCatchEvent *intermediate_catch_event.IntermediateCatchEvent
		intermediateCatchEvent, err = intermediate_catch_event.NewIntermediateCatchEvent(process.Element,
			process.Definitions, element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup,
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
		var aTask *task.Task
		aTask, err = task.NewTask(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
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
		var exclusiveGateway *exclusive_gateway.ExclusiveGateway
		exclusiveGateway, err = exclusive_gateway.NewExclusiveGateway(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, exclusiveGateway)
		if err != nil {
			return
		}
	}

	for i := range *process.Element.ParallelGateways() {
		element := &(*process.Element.ParallelGateways())[i]
		var parallelGateway *parallel_gateway.ParallelGateway
		parallelGateway, err = parallel_gateway.NewParallelGateway(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
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
		var eventBasedGateway *event_based_gateway.EventBasedGateway
		eventBasedGateway, err = event_based_gateway.NewEventBasedGateway(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
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
