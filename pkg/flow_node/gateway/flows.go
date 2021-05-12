package gateway

import (
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/sequence_flow"
)

func DistributeFlows(awaitingActions []chan flow_node.Action, sequenceFlows []*sequence_flow.SequenceFlow) {
	indices := make([]int, len(sequenceFlows))
	for i := range indices {
		indices[i] = i
	}

	for i, action := range awaitingActions {
		rangeEnd := i + 1

		// If this is a last channel awaiting action
		if rangeEnd == len(awaitingActions) {
			// give it the remainder of sequence flows
			rangeEnd = len(sequenceFlows)
		}

		if rangeEnd <= len(sequenceFlows) {
			action <- flow_node.FlowAction{
				SequenceFlows:      sequenceFlows[i:rangeEnd],
				UnconditionalFlows: indices[0 : rangeEnd-i],
			}
		} else {
			// signal completion to flows that aren't
			// getting any flows
			action <- flow_node.CompleteAction{}
		}
	}
}
