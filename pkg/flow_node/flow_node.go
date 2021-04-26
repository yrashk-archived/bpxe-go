package flow_node

import (
	"fmt"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

type FlowNode struct {
	Id           bpmn.Id
	Definitions  *bpmn.Definitions
	Incoming     []sequence_flow.SequenceFlow
	Outgoing     []sequence_flow.SequenceFlow
	EventIngress event.ProcessEventConsumer
	EventEgress  event.ProcessEventSource
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
	flowNode *bpmn.FlowNode,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *FlowNode, err error) {
	incoming, err := sequenceFlows(process, definitions, flowNode.Incomings())
	if err != nil {
		return
	}
	outgoing, err := sequenceFlows(process, definitions, flowNode.Outgoings())
	if err != nil {
		return
	}
	var ownId string
	if ownIdPtr, present := flowNode.Id(); !present {
		err = errors.NotFoundError{
			Expected: fmt.Sprintf("flow node %#v to have an ID", flowNode),
		}
		return
	} else {
		ownId = *ownIdPtr
	}
	node = &FlowNode{
		Id:              ownId,
		Definitions:     definitions,
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
