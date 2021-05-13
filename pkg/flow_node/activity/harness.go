package activity

import (
	"sync"
	"sync/atomic"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/event/catch_event"
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

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type Harness struct {
	flow_node.FlowNode
	element         bpmn.FlowNodeInterface
	runnerChannel   chan message
	activity        Activity
	activeBoundary  <-chan bool
	active          int32
	idGenerator     id.Generator
	instanceBuilder event.InstanceBuilder
	cancellation    sync.Once
	lock            sync.RWMutex
	eventConsumers  []event.ProcessEventConsumer
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

type Constructor = func(process *bpmn.Process,
	definitions *bpmn.Definitions,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node Activity, err error)

func NewHarness(process *bpmn.Process,
	definitions *bpmn.Definitions,
	element *bpmn.FlowNode,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
	idGenerator id.Generator,
	constructor Constructor,
	instanceBuilder event.InstanceBuilder,
) (node *Harness, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		element,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	var activity Activity
	activity, err = constructor(
		process,
		definitions,
		eventIngress,
		eventEgress,
		tracer,
		flowNodeMapping,
		flowWaitGroup,
	)
	if err != nil {
		return
	}

	boundaryEvents := make([]*bpmn.BoundaryEvent, 0)

	for i := range *process.BoundaryEvents() {
		boundaryEvent := &(*process.BoundaryEvents())[i]
		if *boundaryEvent.AttachedToRef() == flowNode.Id {
			boundaryEvents = append(boundaryEvents, boundaryEvent)
		}
	}

	node = &Harness{
		FlowNode:        *flowNode,
		element:         element,
		runnerChannel:   make(chan message, len(flowNode.Incoming)*2+1),
		activity:        activity,
		activeBoundary:  activity.ActiveBoundary(),
		idGenerator:     idGenerator,
		instanceBuilder: instanceBuilder,
	}

	err = node.EventEgress.RegisterProcessEventConsumer(node)
	if err != nil {
		return
	}

	for i := range boundaryEvents {
		boundaryEvent := boundaryEvents[i]
		catchEvent, err := catch_event.NewCatchEvent(node.Process, node.Definitions, &boundaryEvent.CatchEvent,
			node.EventIngress, node, node.Tracer, node.FlowNodeMapping, node.FlowWaitGroup, node.instanceBuilder)
		if err != nil {
			node.Tracer.Trace(tracing.ErrorTrace{Error: err})
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
			newFlow := flow.NewFlow(node.FlowNode.Definitions, catchEvent, node.FlowNode.Tracer,
				node.FlowNode.FlowNodeMapping, node.FlowNode.FlowWaitGroup, node.idGenerator, actionTransformer)
			newFlow.Start()
		}
	}
	go node.runner()
	return
}

func (node *Harness) runner() {
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
			case incomingMessage:
				node.activity.Incoming(m.index)
			case nextActionMessage:
				m.response <- node.activity.NextAction(m.flow)
			default:
			}
		}
	}
}

func (node *Harness) NextAction(flow flow_interface.T) chan flow_node.Action {
	response := make(chan chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{flow: flow, response: response}
	return <-response
}

func (node *Harness) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *Harness) Element() bpmn.FlowNodeInterface {
	return node.element
}
