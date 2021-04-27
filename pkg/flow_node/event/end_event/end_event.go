package end_event

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/id"
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
	startEventsActivated []*bpmn.StartEvent
}

func NewEndEvent(process *bpmn.Process,
	definitions *bpmn.Definitions,
	endEvent *bpmn.EndEvent,
	eventIngress events.ProcessEventConsumer,
	eventEgress events.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *EndEvent, err error) {
	flowNodePtr, err := flow_node.NewFlowNode(
		process,
		definitions,
		&endEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
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
		startEventsActivated: make([]*bpmn.StartEvent, 0),
	}
	go node.runner()
	return
}

func (node *EndEvent) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case incomingMessage:
			node.activated = true
		case nextActionMessage:
			// If the node hasn't been activated, it's too early
			if !node.activated {
				m.response <- flow_node.NoAction{}
				continue
			}
			// If the node already completed, then we essentially fuse it
			if node.completed {
				m.response <- flow_node.CompleteAction{}
				continue
			}

			if _, err := node.FlowNode.EventIngress.ConsumeProcessEvent(
				events.MakeEndEvent(node.element),
			); err == nil {
				node.completed = true
				m.response <- flow_node.CompleteAction{}
			} else {
				node.FlowNode.Tracer.Trace(tracing.ErrorTrace{Error: err})
			}
		default:
		}
	}
}

func (node *EndEvent) NextAction(id.Id) flow_node.Action {
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
