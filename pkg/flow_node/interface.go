package flow_node

import (
	"bpxe.org/pkg/bpmn"
)

type FlowNodeInterface interface {
	Outgoing
	Incoming
	Element() bpmn.FlowNodeInterface
}
