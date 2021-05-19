// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package event

import (
	"bpxe.org/pkg/bpmn"
)

// Instance is a unifying interface for representing event definition within
// an execution context (useful for event definitions like timer, condition, etc.)
type Instance interface {
}

// definitionInstance is a simple wrapper for bpmn.EventDefinitionInterface
// that adds no extra context
type definitionInstance struct {
	bpmn.EventDefinitionInterface
}

// NewInstance is a default event instance builder that creates Instance simply by
// enclosing bpmn.EventDefinitionInterface
func NewInstance(def bpmn.EventDefinitionInterface) Instance {
	return &definitionInstance{EventDefinitionInterface: def}
}

// InstanceBuilder allows supplying custom instance builders that interact with the
// rest of the system and add context for further matching
type InstanceBuilder interface {
	NewEventInstance(def bpmn.EventDefinitionInterface) Instance
}

type DefaultInstanceBuilder struct{}

func (d DefaultInstanceBuilder) NewEventInstance(def bpmn.EventDefinitionInterface) Instance {
	return NewInstance(def)
}
