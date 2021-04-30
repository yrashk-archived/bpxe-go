package activity

import (
	"bpxe.org/pkg/flow_node"
)

// Activity is a generic interface to flow nodes that are activities
type Activity interface {
	flow_node.FlowNodeInterface
	// ActiveBoundary returns a channel that signals `true` when activity becomes
	// active and, subsequently, `false` when it's over.
	ActiveBoundary() <-chan bool
	// Cancel initiates a cancellation of activity and returns a channel
	// that will signal a boolean (`true` if cancellation was successful,
	// `false` otherwise)
	Cancel() <-chan bool
}
