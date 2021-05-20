// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
)

type EventInstantiationAttemptedTrace struct {
	Event   event.Event
	Element bpmn.FlowNodeInterface
}

func (e EventInstantiationAttemptedTrace) TraceInterface() {}
