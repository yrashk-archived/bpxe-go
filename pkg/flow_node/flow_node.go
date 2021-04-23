package flow_node

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/events"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

type FlowNode struct {
	Incoming     []sequence_flow.SequenceFlow
	Outgoing     []sequence_flow.SequenceFlow
	EventIngress events.ProcessEventConsumer
	EventEgress  events.ProcessEventSource
	Tracer       *tracing.Tracer
	Process      *bpmn.Process
	*FlowNodeMapping
	FlowWaitGroup *sync.WaitGroup
}

func sequenceFlows(process *bpmn.Process,
	definitions *bpmn.Definitions,
	flows *[]bpmn.QName) (result []sequence_flow.SequenceFlow, err error) {
	result = make([]sequence_flow.SequenceFlow, len(*flows))
	for i := range result {
		identifier := (*flows)[i]
		exactId := bpmn.ExactId(identifier)
		if element, found := process.FindBy(func(e bpmn.Element) bool {
			_, ok := e.(*bpmn.SequenceFlow)
			return ok && exactId(e)
		}); found {
			result[i] = sequence_flow.MakeSequenceFlow(element.(*bpmn.SequenceFlow), definitions)
		} else {
			err = errors.NotFoundError{Expected: identifier}
			return
		}
	}
	return
}

func NewFlowNode(process *bpmn.Process,
	definitions *bpmn.Definitions,
	flow_node *bpmn.FlowNode,
	eventIngress events.ProcessEventConsumer,
	eventEgress events.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *FlowNode, err error) {
	incoming, err := sequenceFlows(process, definitions, flow_node.Incomings())
	if err != nil {
		return
	}
	outgoing, err := sequenceFlows(process, definitions, flow_node.Outgoings())
	if err != nil {
		return
	}
	node = &FlowNode{
		Incoming:        incoming,
		Outgoing:        outgoing,
		EventIngress:    eventIngress,
		EventEgress:     eventEgress,
		Tracer:          tracer,
		Process:         process,
		FlowNodeMapping: flowNodeMapping,
		FlowWaitGroup:   flowWaitGroup,
	}
	return
}
