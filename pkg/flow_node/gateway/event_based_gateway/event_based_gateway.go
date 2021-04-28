package event_based_gateway

import (
	"sync"
	"sync/atomic"

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
	flowId   id.Id
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type EventBasedGateway struct {
	flow_node.FlowNode
	element       *bpmn.EventBasedGateway
	runnerChannel chan message
	activated     bool
}

func NewEventBasedGateway(process *bpmn.Process, definitions *bpmn.Definitions, eventBasedGateway *bpmn.EventBasedGateway,
	eventIngress event.ProcessEventConsumer, eventEgress event.ProcessEventSource, tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping, flowWaitGroup *sync.WaitGroup) (node *EventBasedGateway, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&eventBasedGateway.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}

	node = &EventBasedGateway{
		FlowNode:      *flowNode,
		element:       eventBasedGateway,
		runnerChannel: make(chan message),
		activated:     false,
	}
	go node.runner()
	return
}

func (node *EventBasedGateway) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case incomingMessage:
			node.activated = true
		case nextActionMessage:
			if node.activated {
				var first int32 = 0
				terminate := make(flow_node.Terminate)
				m.response <- flow_node.FlowAction{
					Terminate:     terminate,
					SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing),
					ActionTransformer: func(flowId id.Id, action flow_node.Action) flow_node.Action {
						// only first one is to flow
						if atomic.CompareAndSwapInt32(&first, 0, 1) {
							node.Tracer.Trace(DeterminationMadeTrace{Element: node.element})
							test := func(anotherflowId id.Id) bool {
								// don't terminate the first (successful) flow
								return anotherflowId != flowId
							}
							// Send a termination for every outgoing flow
							for range node.Outgoing {
								terminate <- test
							}
							return action
						} else {
							return flow_node.CompleteAction{}
						}
					},
				}
			} else {
				m.response <- flow_node.NoAction{}
			}
		default:
		}
	}
}

func (node *EventBasedGateway) NextAction(flowId id.Id) flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flowId: flowId}
	return <-response
}

func (node *EventBasedGateway) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *EventBasedGateway) Element() bpmn.FlowNodeInterface {
	return node.element
}
