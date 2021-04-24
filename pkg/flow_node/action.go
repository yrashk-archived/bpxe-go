package flow_node

import (
	"bpxe.org/pkg/sequence_flow"
)

type Action interface {
	action()
}

type ProbeAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
	// Channel that will be used to receive an array
	// of sequence flow indices that have successful
	// condition expressions (or none)
	ProbeListener chan []int
}

func (action ProbeAction) action() {}

type FlowAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
	// Index of sequence flows that should flow without
	// conditionExpression being evaluated
	UnconditionalFlows []int
}

func (action FlowAction) action() {}

type CompleteAction struct{}

func (action CompleteAction) action() {}

type NoAction struct{}

func (action NoAction) action() {}
