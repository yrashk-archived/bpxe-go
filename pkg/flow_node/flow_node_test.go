package flow_node

import (
	"encoding/xml"
	"sync"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"
)

var defaultDefinitions = bpmn.DefaultDefinitions()

func TestNewFlowNode(t *testing.T) {
	var sampleDoc bpmn.Definitions
	var err error
	sample, err := testdata.ReadFile("testdata/sample.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(sample, &sampleDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}

	var waitGroup sync.WaitGroup
	if proc, found := sampleDoc.FindBy(bpmn.ExactId("sample")); found {
		if flowNode, found := sampleDoc.FindBy(bpmn.ExactId("either")); found {
			node, err := NewFlowNode(proc.(*bpmn.Process),
				&defaultDefinitions,
				&flowNode.(*bpmn.ParallelGateway).FlowNode,
				event.VoidProcessEventConsumer{},
				event.VoidProcessEventSource{},
				tracing.NewTracer(), NewLockedFlowNodeMapping(),
				&waitGroup,
			)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(node.Incoming))
			t.Logf("%+v", node.Incoming[0])
			if incomingSeqFlowId, present := node.Incoming[0].Id(); present {
				assert.Equal(t, *incomingSeqFlowId, "x1")
			} else {
				t.Fatalf("Sequence flow x1 has no matching ID")
			}
			assert.Equal(t, 2, len(node.Outgoing))
			if outgoingSeqFlowId, present := node.Outgoing[0].Id(); present {
				assert.Equal(t, *outgoingSeqFlowId, "x2")
			} else {
				t.Fatalf("Sequence flow x2 has no matching ID")
			}
			if outgoingSeqFlowId, present := node.Outgoing[1].Id(); present {
				assert.Equal(t, *outgoingSeqFlowId, "x3")
			} else {
				t.Fatalf("Sequence flow x3 has no matching ID")
			}
		} else {
			t.Fatalf("Can't find flow node `either`")
		}
	} else {
		t.Fatalf("Can't find process `sample`")
	}
}
