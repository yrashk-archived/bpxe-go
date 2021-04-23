package start_event

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/flow"
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

type StartEvent struct {
	flow_node.FlowNode
	element       *bpmn.StartEvent
	runnerChannel chan message
	activated     bool
}

func NewStartEvent(process *bpmn.Process,
	definitions *bpmn.Definitions,
	startEvent *bpmn.StartEvent,
	eventIngress events.ProcessEventConsumer,
	eventEgress events.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *StartEvent, err error) {
	flow_node, err := flow_node.NewFlowNode(process,
		definitions,
		&startEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	node = &StartEvent{
		FlowNode:      *flow_node,
		element:       startEvent,
		runnerChannel: make(chan message),
		activated:     false,
	}
	go node.runner()
	err = node.EventEgress.RegisterProcessEventConsumer(node)
	if err != nil {
		return
	}
	return
}

func (node *StartEvent) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case nextActionMessage:
			if !node.activated {
				node.activated = true
				m.response <- flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
			} else {
				m.response <- flow_node.CompleteAction{}
			}
		default:
		}
	}
}

func (node *StartEvent) ConsumeProcessEvent(
	ev events.ProcessEvent,
) (result events.EventConsumptionResult, err error) {
	switch ev.(type) {
	case *events.StartEvent:
		newFlow := flow.NewFlow(node.FlowNode.Definitions, node, node.FlowNode.Tracer,
			node.FlowNode.FlowNodeMapping, node.FlowNode.FlowWaitGroup)
		newFlow.Start()
	default:
	}
	result = events.EventConsumed
	return
}

func (node *StartEvent) NextAction() flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return <-response
}

func (node *StartEvent) Incoming(int) {
	// Do nothing, there are no incoming flows for start events
	// but we have to implement it to satisfy FlowNodeInterface
}

func (node *StartEvent) Element() bpmn.FlowNodeInterface {
	return node.element
}
