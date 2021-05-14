// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package flow_node

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/sequence_flow"
)

type Action interface {
	action()
}

type ProbeAction struct {
	SequenceFlows []*sequence_flow.SequenceFlow
	// ProbeReport is a function that needs to be called
	// wth sequence flow indices that have successful
	// condition expressions (or none)
	ProbeReport func([]int)
}

func (action ProbeAction) action() {}

type ActionTransformer func(sequenceFlowId *bpmn.IdRef, action Action) Action
type Terminate func(sequenceFlowId *bpmn.IdRef) chan bool

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
