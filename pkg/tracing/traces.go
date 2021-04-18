package tracing

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/sequence_flow"
)

// Interface for actual data traces
type Trace interface {
	TraceInterface()
}

type FlowTrace struct {
	Source        bpmn.FlowNodeInterface
	SequenceFlows []*sequence_flow.SequenceFlow
}

func (t FlowTrace) TraceInterface() {}

type CompletionTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t CompletionTrace) TraceInterface() {}

type ErrorTrace struct {
	Error error
}

func (t ErrorTrace) TraceInterface() {}
