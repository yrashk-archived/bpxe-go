package flow_node

import (
	"bpxe.org/pkg/sequence_flow"
)

type Action interface {
	action()
}

type ProbeAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
}

func (action ProbeAction) action() {}

type FlowAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
}

func (action FlowAction) action() {}

type CompleteAction struct{}

func (action CompleteAction) action() {}

type NoAction struct{}

func (action NoAction) action() {}
