package activity

import (
	"sync"

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
	element                   bpmn.FlowNodeInterface
	runnerChannel             chan message
	activity                  Activity
	activeBoundary            <-chan bool
	active                    bool
	idGenerator               id.Generator
	boundaryEvents            []*bpmn.BoundaryEvent
	boundaryEventTerminations []chan bool
	instanceBuilder           event.InstanceBuilder
	cancellation              sync.Once
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
	boundaryEventsTerminations := make([]chan bool, 0)

	for i := range *process.BoundaryEvents() {
		boundaryEvent := &(*process.BoundaryEvents())[i]
		if *boundaryEvent.AttachedToRef() == flowNode.Id {
			boundaryEvents = append(boundaryEvents, boundaryEvent)
			boundaryEventsTerminations = append(boundaryEventsTerminations, make(chan bool))
		}
	}

	node = &Harness{
		FlowNode:                  *flowNode,
		element:                   element,
		runnerChannel:             make(chan message),
		activity:                  activity,
		activeBoundary:            activity.ActiveBoundary(),
		idGenerator:               idGenerator,
		boundaryEvents:            boundaryEvents,
		boundaryEventTerminations: boundaryEventsTerminations,
		instanceBuilder:           instanceBuilder,
	}
	go node.runner()
	return
}

func (node *Harness) runner() {
	boundaryEventFlows := make([]*flow.Flow, len(node.boundaryEvents))
	for {
		select {
		case activeBoundary := <-node.activeBoundary:
			if activeBoundary && !node.active {
				// Opening active boundary
				node.Tracer.Trace(ActiveBoundaryTrace{Start: activeBoundary, Node: node.activity.Element()})
				for i := range node.boundaryEvents {
					boundaryEvent := node.boundaryEvents[i]
					catchEvent, err := catch_event.NewCatchEvent(node.Process, node.Definitions, &boundaryEvent.CatchEvent,
						node.EventIngress, node.EventEgress, node.Tracer, node.FlowNodeMapping, node.FlowWaitGroup, node.instanceBuilder)
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
						newFlow.SetTerminate(func(*bpmn.IdRef) chan bool {
							return node.boundaryEventTerminations[i]
						})
						boundaryEventFlows[i] = newFlow
						newFlow.Start()
					}
				}
			} else if !activeBoundary && node.active {
				// Closing active boundary
				node.Tracer.Trace(ActiveBoundaryTrace{Start: activeBoundary, Node: node.activity.Element()})
				// Terminate boundary events
				for i := range node.boundaryEventTerminations {
					select {
					// attempt to send if waiting
					case node.boundaryEventTerminations[i] <- true:
					default:
						// bail otherwise
					}
				}
			}
			node.active = activeBoundary
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
