// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

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
