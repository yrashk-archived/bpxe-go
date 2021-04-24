package flow_node

import (
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
)

type Outgoing interface {
	NextAction(flowId id.Id) Action
}

func AllSequenceFlows(
	sequenceFlows *[]sequence_flow.SequenceFlow,
	exclusion ...func(*sequence_flow.SequenceFlow) bool,
) (result []*sequence_flow.SequenceFlow) {
	result = make([]*sequence_flow.SequenceFlow, 0)
sequenceFlowsLoop:
	for i := range *sequenceFlows {
		for _, exclFun := range exclusion {
			if exclFun(&(*sequenceFlows)[i]) {
				continue sequenceFlowsLoop
			}
		}
		result = append(result, &(*sequenceFlows)[i])
	}
	return
}
