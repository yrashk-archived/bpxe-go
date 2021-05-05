package flow

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
)

type NewFlowTrace struct {
	FlowId id.Id
}

func (t NewFlowTrace) TraceInterface() {}

type FlowTrace struct {
	Source bpmn.FlowNodeInterface
	Flows  []Snapshot
}

func (t FlowTrace) TraceInterface() {}

type FlowTerminationTrace struct {
	FlowId id.Id
	Source bpmn.FlowNodeInterface
}

func (t FlowTerminationTrace) TraceInterface() {}

type CompletionTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t CompletionTrace) TraceInterface() {}

type CeaseFlowTrace struct{}

func (t CeaseFlowTrace) TraceInterface() {}

type VisitTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t VisitTrace) TraceInterface() {}
