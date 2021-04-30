package activity

import (
	"bpxe.org/pkg/bpmn"
)

type ActiveBoundaryTrace struct {
	Start bool
	Node  bpmn.FlowNodeInterface
}

func (b ActiveBoundaryTrace) TraceInterface() {}
