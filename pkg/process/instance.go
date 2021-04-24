package process

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity/task"
	"bpxe.org/pkg/flow_node/event/end_event"
	"bpxe.org/pkg/flow_node/event/start_event"
	"bpxe.org/pkg/tracing"
)

type ProcessInstance struct {
	process         *Process
	eventConsumers  []events.ProcessEventConsumer
	Tracer          *tracing.Tracer
	flowNodeMapping *flow_node.FlowNodeMapping
	flowWaitGroup   sync.WaitGroup
	complete        sync.RWMutex
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
			element, instance, instance, tracer, instance.flowNodeMapping, &instance.flowWaitGroup)
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
	lockChan := make(chan bool)
	go func() {
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
		traces := instance.Tracer.Subscribe()

		// So, at first, we wait for (1.1) to occur
		// [(1.2) will be addded when we actually support them]

		for {
			if len(startEventsActivated) == len(*instance.process.Element.StartEvents()) {
				break
			}

			trace := <-traces

			switch t := trace.(type) {
			case tracing.FlowTerminationTrace:
				switch flowNode := t.Source.(type) {
				case *bpmn.StartEvent:
					startEventsActivated = append(startEventsActivated, flowNode)
				default:
				}
			case tracing.FlowTrace:
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
		instance.Tracer.Trace(tracing.CeaseFlowTrace{})
		instance.complete.Unlock()
	}()
	<-lockChan
	close(lockChan)
	event := events.MakeStartEvent()
	_, err = instance.ConsumeProcessEvent(&event)
	if err != nil {
		return
	}
	return
}

// Waits until the instance is complete. Returns true if the instance was complete,
// false if the context signalled `Done`
func (instance *ProcessInstance) WaitUntilComplete(ctx context.Context) (complete bool) {
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
