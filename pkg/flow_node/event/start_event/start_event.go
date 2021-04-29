package start_event

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
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

type StartEvent struct {
	flow_node.FlowNode
	element       *bpmn.StartEvent
	runnerChannel chan message
	activated     bool
	idGenerator   id.Generator
}

func NewStartEvent(process *bpmn.Process,
	definitions *bpmn.Definitions,
	startEvent *bpmn.StartEvent,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
	idGenerator id.Generator,
) (node *StartEvent, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&startEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	node = &StartEvent{
		FlowNode:      *flowNode,
		element:       startEvent,
		runnerChannel: make(chan message),
		activated:     false,
		idGenerator:   idGenerator,
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
	ev event.ProcessEvent,
) (result event.ConsumptionResult, err error) {
	switch ev.(type) {
	case *event.StartEvent:
		newFlow := flow.NewFlow(node.FlowNode.Definitions, node, node.FlowNode.Tracer,
			node.FlowNode.FlowNodeMapping, node.FlowNode.FlowWaitGroup, node.idGenerator, nil)
		newFlow.Start()
	default:
	}
	result = event.Consumed
	return
}

func (node *StartEvent) NextAction(id.Id) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *StartEvent) Incoming(int) {
	// Do nothing, there are no incoming flows for start events
	// but we have to implement it to satisfy FlowNodeInterface
}

func (node *StartEvent) Element() bpmn.FlowNodeInterface {
	return node.element
}
