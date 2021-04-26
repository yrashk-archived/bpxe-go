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
