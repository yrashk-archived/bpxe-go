package end_event

import (
	"testing"

	"bpxe.org/pkg/flow_node"
)

func TestEndEventInterface(t *testing.T) {
	var _ flow_node.FlowNodeInterface = &EndEvent{}
}
