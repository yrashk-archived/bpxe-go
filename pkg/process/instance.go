package process

import (
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/event/end_event"
	"bpxe.org/pkg/flow_node/event/start_event"
	"bpxe.org/pkg/tracing"
)

type ProcessInstance struct {
	process         *Process
	eventConsumers  []events.ProcessEventConsumer
	Tracer          *tracing.Tracer
	flowNodeMapping *flow_node.FlowNodeMapping
}

func NewProcessInstance(process *Process) (instance *ProcessInstance, err error) {
	eventConsumers := make([]events.ProcessEventConsumer, 0)
	tracer := tracing.NewTracer()
	instance = &ProcessInstance{
		process:         process,
		eventConsumers:  eventConsumers,
		Tracer:          tracer,
		flowNodeMapping: flow_node.NewLockedFlowNodeMapping(),
	}
	for i := range *process.Element.StartEvents() {
		element := &(*process.Element.StartEvents())[i]
		var startEvent *start_event.StartEvent
		startEvent, err = start_event.NewStartEvent(process.Element, process.Definitions,
			element, instance, instance, tracer, instance.flowNodeMapping)
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
			element, instance, instance, tracer, instance.flowNodeMapping)
		if err != nil {
			return
		}
		err = instance.flowNodeMapping.RegisterElementToFlowNode(element, endEvent)
		if err != nil {
			return
		}
	}

	instance.flowNodeMapping.Finalize()

	return
}

func (instance *ProcessInstance) ConsumeProcessEvent(ev events.ProcessEvent) (result events.EventConsumptionResult, err error) {
	result, err = events.ForwardProcessEvent(ev, &instance.eventConsumers)
	return
}

func (instance *ProcessInstance) RegisterProcessEventConsumer(ev events.ProcessEventConsumer) (err error) {
	instance.eventConsumers = append(instance.eventConsumers, ev)
	return
}

func (instance *ProcessInstance) Run() (err error) {
	event := events.MakeStartEvent()
	_, err = instance.ConsumeProcessEvent(&event)
	if err != nil {
		return
	}
	return
}
