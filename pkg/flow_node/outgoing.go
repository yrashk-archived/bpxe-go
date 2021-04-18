package flow_node

import (
	"bpxe.org/pkg/sequence_flow"
)

type Outgoing interface {
	NextAction() Action
}

func AllSequenceFlows(
	sequenceFlows *[]sequence_flow.SequenceFlow,
) (result []*sequence_flow.SequenceFlow) {
	result = make([]*sequence_flow.SequenceFlow, len(*sequenceFlows))
	for i := range *sequenceFlows {
		result[i] = &(*sequenceFlows)[i]
	}
	return
}
