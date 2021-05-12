package flow_interface

import (
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
)

// T specifies an interface for BPMN flows
type T interface {
	// Id returns flow's unique identifier
	Id() id.Id
	// SequenceFlow returns an inbound sequence flow this flow
	// is currently at.
	SequenceFlow() *sequence_flow.SequenceFlow
}
