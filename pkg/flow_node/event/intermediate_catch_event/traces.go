package intermediate_catch_event

import (
	"bpxe.org/pkg/bpmn"
)

type ActiveListeningTrace struct {
	Node *bpmn.IntermediateCatchEvent
}

func (t ActiveListeningTrace) TraceInterface() {}
