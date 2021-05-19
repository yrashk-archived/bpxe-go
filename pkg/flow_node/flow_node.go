// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

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

// Wiring holds all necessary "wiring" for functioning of
// flow nodes: definitions, process, sequence flow, event management,
// tracer, flow node mapping and a flow wait group
type Wiring struct {
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
			result[i] = sequence_flow.Make(element.(*bpmn.SequenceFlow), definitions)
		} else {
			err = errors.NotFoundError{Expected: identifier}
			return
		}
	}
	return
}

func New(process *bpmn.Process,
	definitions *bpmn.Definitions,
	flowNode *bpmn.FlowNode,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *Wiring, err error) {
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
	node = &Wiring{
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

// CloneFor copies receiver, overriding Id, Incoming, Outgoing for a given flowNode
func (wiring *Wiring) CloneFor(flowNode *bpmn.FlowNode) (result *Wiring, err error) {
	incoming, err := sequenceFlows(wiring.Process, wiring.Definitions, flowNode.Incomings())
	if err != nil {
		return
	}
	outgoing, err := sequenceFlows(wiring.Process, wiring.Definitions, flowNode.Outgoings())
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
	result = &Wiring{
		Id:              ownId,
		Definitions:     wiring.Definitions,
		Incoming:        incoming,
		Outgoing:        outgoing,
		EventIngress:    wiring.EventIngress,
		EventEgress:     wiring.EventEgress,
		Tracer:          wiring.Tracer,
		Process:         wiring.Process,
		FlowNodeMapping: wiring.FlowNodeMapping,
		FlowWaitGroup:   wiring.FlowWaitGroup,
	}
	return
}
