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

// DefinitionInstance is a unifying interface for representing event definition within
// an execution context (useful for event definitions like timer, condition, etc.)
type DefinitionInstance interface {
	EventDefinition() bpmn.EventDefinitionInterface
}

// wrappedDefinitionInstance is a simple wrapper for bpmn.EventDefinitionInterface
// that adds no extra context
type wrappedDefinitionInstance struct {
	definition bpmn.EventDefinitionInterface
}

func (d *wrappedDefinitionInstance) EventDefinition() bpmn.EventDefinitionInterface {
	return d.definition
}

// WrapEventDefinition is a default event instance builder that creates Instance simply by
// enclosing bpmn.EventDefinitionInterface
func WrapEventDefinition(def bpmn.EventDefinitionInterface) DefinitionInstance {
	return &wrappedDefinitionInstance{definition: def}
}

// DefinitionInstanceBuilder allows supplying custom instance builders that interact with the
// rest of the system and add context for further matching
type DefinitionInstanceBuilder interface {
	NewEventInstance(def bpmn.EventDefinitionInterface) DefinitionInstance
}

type wrappingDefinitionInstanceBuilder struct{}

var WrappingDefinitionInstanceBuilder = wrappingDefinitionInstanceBuilder{}

func (d wrappingDefinitionInstanceBuilder) NewEventInstance(def bpmn.EventDefinitionInterface) DefinitionInstance {
	return WrapEventDefinition(def)
}
