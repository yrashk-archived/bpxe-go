package start_event

import (
	"testing"

	"bpxe.org/pkg/flow_node"
)

func TestStartEventInterface(t *testing.T) {
	var _ flow_node.FlowNodeInterface = &StartEvent{}
}
