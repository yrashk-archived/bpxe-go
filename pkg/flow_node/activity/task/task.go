package task

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
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

type Task struct {
	flow_node.FlowNode
	element       *bpmn.Task
	runnerChannel chan message
	activated     bool
}

func NewTask(process *bpmn.Process,
	definitions *bpmn.Definitions,
	startEvent *bpmn.Task,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *Task, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&startEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	node = &Task{
		FlowNode:      *flowNode,
		element:       startEvent,
		runnerChannel: make(chan message),
		activated:     false,
	}
	go node.runner()
	return
}

func (node *Task) runner() {
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

func (node *Task) NextAction(id.Id) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Task) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *Task) Element() bpmn.FlowNodeInterface {
	return node.element
}
