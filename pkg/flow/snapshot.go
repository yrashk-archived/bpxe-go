package flow

import (
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
)

type Snapshot struct {
	flowId       id.Id
	sequenceFlow *sequence_flow.SequenceFlow
}

func (s *Snapshot) Id() id.Id {
	return s.flowId
}

func (s *Snapshot) SequenceFlow() *sequence_flow.SequenceFlow {
	return s.sequenceFlow
}
