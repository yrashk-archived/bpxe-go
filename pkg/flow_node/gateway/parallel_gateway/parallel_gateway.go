package parallel_gateway

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
	flowId   id.Id
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type ParallelGateway struct {
	flow_node.FlowNode
	element               *bpmn.ParallelGateway
	runnerChannel         chan message
	reportedIncomingFlows []int
	awaitingActions       []chan flow_node.Action
	noOfIncomingFlows     int
}

func NewParallelGateway(process *bpmn.Process,
	definitions *bpmn.Definitions,
	parallelGateway *bpmn.ParallelGateway,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *ParallelGateway, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&parallelGateway.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}

	node = &ParallelGateway{
		FlowNode:              *flowNode,
		element:               parallelGateway,
		runnerChannel:         make(chan message),
		reportedIncomingFlows: make([]int, 0),
		awaitingActions:       make([]chan flow_node.Action, 0),
		noOfIncomingFlows:     len(flowNode.Incoming),
	}
	go node.runner()
	if err != nil {
		return
	}
	return
}

func (node *ParallelGateway) flowWhenReady() {
	if len(node.reportedIncomingFlows) == node.noOfIncomingFlows &&
		len(node.awaitingActions) == node.noOfIncomingFlows {
		node.reportedIncomingFlows = make([]int, 0)
		awaitingActions := node.awaitingActions
		node.awaitingActions = make([]chan flow_node.Action, 0)
		sequenceFlows := flow_node.AllSequenceFlows(&node.Outgoing)
		indices := make([]int, len(sequenceFlows))
		for i := range indices {
			indices[i] = i
		}

		for i, action := range awaitingActions {
			rangeEnd := i + 1

			// If this is a last channel awaiting action
			if rangeEnd == len(awaitingActions) {
				// give it the remainder of sequence flows
				rangeEnd = len(sequenceFlows)
			}

			if rangeEnd <= len(sequenceFlows) {
				action <- flow_node.FlowAction{
					SequenceFlows:      sequenceFlows[i:rangeEnd],
					UnconditionalFlows: indices[0 : rangeEnd-i],
				}
			} else {
				// signal completion to flows that aren't
				// getting any flows
				action <- flow_node.CompleteAction{}
			}
		}
	}

}

func (node *ParallelGateway) runner() {
loop:
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case incomingMessage:
			for _, incomingIndex := range node.reportedIncomingFlows {
				if m.index == incomingIndex {
					continue loop
				}
			}
			node.reportedIncomingFlows = append(node.reportedIncomingFlows, m.index)
			node.flowWhenReady()
		case nextActionMessage:
			node.awaitingActions = append(node.awaitingActions, m.response)
			node.flowWhenReady()
		default:
		}
	}
}

func (node *ParallelGateway) NextAction(flowId id.Id) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flowId: flowId}
	return response
}

func (node *ParallelGateway) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *ParallelGateway) Element() bpmn.FlowNodeInterface {
	return node.element
}
