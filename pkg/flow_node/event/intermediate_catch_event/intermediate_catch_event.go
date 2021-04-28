package intermediate_catch_event

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

type processEventMessage struct {
	event event.ProcessEvent
}

func (m processEventMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type IntermediateCatchEvent struct {
	flow_node.FlowNode
	element         *bpmn.IntermediateCatchEvent
	runnerChannel   chan message
	activated       bool
	awaitingActions []chan flow_node.Action
	eventInstances  []event.Instance
	matchedEvents   []bool
}

func NewIntermediateCatchEvent(process *bpmn.Process, definitions *bpmn.Definitions,
	intermediateCatchEvent *bpmn.IntermediateCatchEvent, eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource, tracer *tracing.Tracer, flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup, instanceBuilder event.InstanceBuilder) (node *IntermediateCatchEvent, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&intermediateCatchEvent.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}
	eventDefinitions := intermediateCatchEvent.EventDefinitions()
	eventInstances := make([]event.Instance, len(eventDefinitions))

	for i, eventDefinition := range eventDefinitions {
		eventInstances[i] = instanceBuilder.NewEventInstance(eventDefinition)
	}

	node = &IntermediateCatchEvent{
		FlowNode:        *flowNode,
		element:         intermediateCatchEvent,
		runnerChannel:   make(chan message),
		activated:       false,
		awaitingActions: make([]chan flow_node.Action, 0),
		eventInstances:  eventInstances,
		matchedEvents:   make([]bool, len(eventDefinitions)),
	}
	go node.runner()
	err = node.EventEgress.RegisterProcessEventConsumer(node)
	if err != nil {
		return
	}
	return
}

func (node *IntermediateCatchEvent) runner() {
loop:
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case processEventMessage:
			if node.activated {
				if len(node.eventInstances) == 0 {
					//lint:ignore SA4006 not sure why it's complaining, `ok` is used
					//nolint:staticcheck
					if _, ok := m.event.(event.NoneEvent); ok {
						goto matched
					}
				} else {
					for i, instance := range node.eventInstances {
						if m.event.MatchesEventInstance(instance) {
							node.matchedEvents[i] = true
							goto matched
						}
					}
				}
				continue loop
			matched:
				for _, matched := range node.matchedEvents {
					if !matched && node.element.ParallelMultiple() {
						continue loop
					}
				}
				awaitingActions := node.awaitingActions
				for _, actionChan := range awaitingActions {
					actionChan <- flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
				}
				node.awaitingActions = make([]chan flow_node.Action, 0)
			}
		case incomingMessage:
			if !node.activated {
				node.activated = true
				node.Tracer.Trace(ActiveListeningTrace{Node: node.element})
			}
		case nextActionMessage:
			if node.activated {
				node.awaitingActions = append(node.awaitingActions, m.response)
			} else {
				m.response <- flow_node.NoAction{}
			}
		default:
		}
	}
}

func (node *IntermediateCatchEvent) ConsumeProcessEvent(
	ev event.ProcessEvent,
) (result event.EventConsumptionResult, err error) {
	node.runnerChannel <- processEventMessage{event: ev}
	result = event.EventConsumed
	return
}

func (node *IntermediateCatchEvent) NextAction(id.Id) flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return <-response
}

func (node *IntermediateCatchEvent) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *IntermediateCatchEvent) Element() bpmn.FlowNodeInterface {
	return node.element
}
