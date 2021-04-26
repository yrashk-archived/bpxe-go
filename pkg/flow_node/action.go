package flow_node

import (
	"bpxe.org/pkg/id"
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

type ActionTransformer func(id.Id, Action) Action
type Terminate chan func(id id.Id) bool

type FlowAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
	// Index of sequence flows that should flow without
	// conditionExpression being evaluated
	UnconditionalFlows []int
	// The actions produced by the targets should be processed by
	// this function
	ActionTransformer
	// If supplied channel sends a function that returns true, the flow action
	// is to be terminated if it wasn't already
	Terminate
}

func (action FlowAction) action() {}

type CompleteAction struct{}

func (action CompleteAction) action() {}

type NoAction struct{}

func (action NoAction) action() {}
