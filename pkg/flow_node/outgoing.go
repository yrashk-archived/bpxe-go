// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package flow_node

import (
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/sequence_flow"
)

type Outgoing interface {
	NextAction(flow flow_interface.T) chan Action
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
