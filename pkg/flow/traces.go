package flow

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
)

type NewFlowTrace struct {
	FlowId id.Id
}

func (t NewFlowTrace) TraceInterface() {}

type FlowTrace struct {
	FlowId        id.Id
	Source        bpmn.FlowNodeInterface
	SequenceFlows []*sequence_flow.SequenceFlow
}

func (t FlowTrace) TraceInterface() {}

type FlowTerminationTrace struct {
	Source bpmn.FlowNodeInterface
}

func (t FlowTerminationTrace) TraceInterface() {}

type CompletionTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t CompletionTrace) TraceInterface() {}

type CeaseFlowTrace struct{}

func (t CeaseFlowTrace) TraceInterface() {}
