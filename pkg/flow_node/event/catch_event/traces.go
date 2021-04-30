package catch_event

import (
	"bpxe.org/pkg/bpmn"
)

type ActiveListeningTrace struct {
	Node *bpmn.CatchEvent
}

func (t ActiveListeningTrace) TraceInterface() {}
