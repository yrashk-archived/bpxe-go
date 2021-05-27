// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package instance

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/tracing"
)

// InstantiationTrace denotes instantiation of a given process
type InstantiationTrace struct {
	InstanceId id.Id
	Process    *bpmn.Process
}

func (i InstantiationTrace) TraceInterface() {}

// Trace wraps any trace with process instance id
type Trace struct {
	InstanceId id.Id
	Trace      tracing.Trace
}

func (t Trace) Unwrap() tracing.Trace {
	return t.Trace
}

func (t Trace) TraceInterface() {}
