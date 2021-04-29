package event_based_gateway

import (
	"fmt"
	"sync"
	"sync/atomic"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
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
		case nextActionMessage:
			var first int32 = 0
			sequenceFlows := flow_node.AllSequenceFlows(&node.Outgoing)
			terminationChannels := make(map[bpmn.IdRef]chan bool)
			for _, sequenceFlow := range sequenceFlows {
				if idPtr, present := sequenceFlow.Id(); present {
					terminationChannels[*idPtr] = make(chan bool)
				} else {
					node.Tracer.Trace(tracing.ErrorTrace{Error: errors.NotFoundError{
						Expected: fmt.Sprintf("id for %#v", sequenceFlow),
					}})
				}
			}
			m.response <- flow_node.FlowAction{
				Terminate: func(sequenceFlowId bpmn.IdRef) chan bool {
					return terminationChannels[sequenceFlowId]
				},
				SequenceFlows: sequenceFlows,
				ActionTransformer: func(sequenceFlowId bpmn.IdRef, action flow_node.Action) flow_node.Action {
					// only first one is to flow
					if atomic.CompareAndSwapInt32(&first, 0, 1) {
						node.Tracer.Trace(DeterminationMadeTrace{Element: node.element})
						for terminationCandidateId, ch := range terminationChannels {
							if terminationCandidateId != sequenceFlowId {
								ch <- true
							}
							close(ch)
						}
						terminationChannels = make(map[bpmn.IdRef]chan bool)
						return action
					} else {
						return flow_node.CompleteAction{}
					}
				},
			}
		default:
		}
	}
}

func (node *EventBasedGateway) NextAction(flowId id.Id) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flowId: flowId}
	return response
}

func (node *EventBasedGateway) Incoming(int) {
}

func (node *EventBasedGateway) Element() bpmn.FlowNodeInterface {
	return node.element
}
