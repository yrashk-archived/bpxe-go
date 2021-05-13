package tests

import (
	"context"
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity"
	"bpxe.org/pkg/flow_node/activity/task"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/assert"

	_ "github.com/stretchr/testify/assert"
)

func TestInterruptingEvent(t *testing.T) {
	testBoundaryEvent(t, "testdata/boundary_event.bpmn", func(visited map[string]bool) {
		assert.False(t, visited["uninterrupted"])
		assert.True(t, visited["interrupted"])
		assert.True(t, visited["end"])
	}, event.NewSignalEvent("sig1"))
}

func TestNonInterruptingEvent(t *testing.T) {
	testBoundaryEvent(t, "testdata/boundary_event.bpmn", func(visited map[string]bool) {
		assert.False(t, visited["interrupted"])
		assert.True(t, visited["uninterrupted"])
		assert.True(t, visited["end"])
	}, event.NewSignalEvent("sig2"))
}

func testBoundaryEvent(t *testing.T, filename string, test func(visited map[string]bool), events ...event.ProcessEvent) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile(filename)
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	ready := make(chan bool)
	if instance, err := proc.Instantiate(); err == nil {
		if node, found := testDoc.FindBy(bpmn.ExactId("task")); found {
			if taskNode, found := instance.FlowNodeMapping().
				ResolveElementToFlowNode(node.(bpmn.FlowNodeInterface)); found {
				harness := taskNode.(*activity.Harness)
				aTask := harness.Activity().(*task.Task)
				aTask.SetBody(func(task *task.Task, ctx context.Context) flow_node.Action {
					select {
					case <-ready:
						return flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&task.FlowNode.Outgoing)}
					case <-ctx.Done():
						return flow_node.CompleteAction{}
					}
				})
			} else {
				t.Fatalf("failed to get the flow node `task`")
			}
		} else {
			t.Fatalf("failed to get the flow node element for `task`")
		}
		// this gives us some room when instance starts up
		traces := instance.Tracer.SubscribeChannel(make(chan tracing.Trace, 32))
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case activity.ActiveBoundaryTrace:
				if id, present := trace.Node.Id(); present && trace.Start {
					if *id == "task" {
						// got to task
						break loop
					}

				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}

		for _, evt := range events {
			_, err = instance.ConsumeProcessEvent(evt)
			assert.Nil(t, err)
		}
		visited := make(map[string]bool)
	loop1:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.VisitTrace:
				if id, present := trace.Node.Id(); present {
					if *id == "uninterrupted" {
						// we're here to we can release the task
						ready <- true
					}
					visited[*id] = true
					if *id == "end" {
						break loop1
					}
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)

		test(visited)

	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}
