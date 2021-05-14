// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package flow

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
)

type NewFlowTrace struct {
	FlowId id.Id
}

func (t NewFlowTrace) TraceInterface() {}

type FlowTrace struct {
	Source bpmn.FlowNodeInterface
	Flows  []Snapshot
}

func (t FlowTrace) TraceInterface() {}

type FlowTerminationTrace struct {
	FlowId id.Id
	Source bpmn.FlowNodeInterface
}

func (t FlowTerminationTrace) TraceInterface() {}

type CompletionTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t CompletionTrace) TraceInterface() {}

type CeaseFlowTrace struct{}

func (t CeaseFlowTrace) TraceInterface() {}

type VisitTrace struct {
	Node bpmn.FlowNodeInterface
}

func (t VisitTrace) TraceInterface() {}
