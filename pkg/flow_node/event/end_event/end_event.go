package end_event

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type EndEvent struct {
	flow_node.FlowNode
	element              *bpmn.EndEvent
	activated            bool
	completed            bool
	eventConsumer        events.ProcessEventConsumer
	runnerChannel        chan message
	traces               chan tracing.Trace
	startEventsActivated []*bpmn.StartEvent
}

func NewEndEvent(process *bpmn.Process,
	definitions *bpmn.Definitions,
	endEvent *bpmn.EndEvent,
	eventIngress events.ProcessEventConsumer,
	eventEgress events.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
) (node *EndEvent, err error) {
	flowNodePtr, err := flow_node.NewFlowNode(
		process,
		definitions,
		&endEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping)
	if err != nil {
		return
	}
	flowNode := *flowNodePtr
	node = &EndEvent{
		FlowNode:             flowNode,
		element:              endEvent,
		activated:            false,
		completed:            false,
		eventConsumer:        eventIngress,
		runnerChannel:        make(chan message),
		traces:               flowNode.Tracer.Subscribe(),
		startEventsActivated: make([]*bpmn.StartEvent, 0),
	}
	go node.runner()
	if err != nil {
		return
	}
	return
}

func (node *EndEvent) runner() {
	for {
		select {
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case incomingMessage:
				node.activated = true
			case nextActionMessage:
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
				// If the node hasn't been activated, it's too early
				if !node.activated {
					m.response <- flow_node.NoAction{}
				}
				// If the node already completed, then we essentially fuse it
				if node.completed {
					m.response <- flow_node.CompleteAction{}
				}
				// To satisfy 1.1, we're observing traces to know
				// when start events were activated. `startEventsActivated`
				// keeps track of these.
				// We're not taking in account remaining tokens (YET).
				if len(node.startEventsActivated) == len(*node.FlowNode.Process.StartEvents()) {
					if _, err := node.FlowNode.EventIngress.ConsumeProcessEvent(
						events.MakeEndEvent(),
					); err == nil {
						node.completed = true
						m.response <- flow_node.CompleteAction{}
					}
				}
			default:
			}
		case trace := <-node.traces:
			switch t := trace.(type) {
			case tracing.FlowTrace:
				for _, sequenceFlow := range t.SequenceFlows {
					if source, err := sequenceFlow.Source(); err == nil {
						switch flowNode := source.(type) {
						case *bpmn.StartEvent:
							node.startEventsActivated = append(node.startEventsActivated, flowNode)
						default:
						}
					}
				}
			default:
			}
		}
	}
}

func (node *EndEvent) NextAction() flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return <-response
}

func (node *EndEvent) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *EndEvent) Element() bpmn.FlowNodeInterface {
	return node.element
}
