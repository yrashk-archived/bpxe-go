package event_based_gateway

import (
	"bpxe.org/pkg/bpmn"
)

type DeterminationMadeTrace struct {
	bpmn.Element
}

func (trace DeterminationMadeTrace) TraceInterface() {}
