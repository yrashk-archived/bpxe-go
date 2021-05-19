// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tests

import (
	"context"
	"testing"

	"bpxe.org/internal"
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity"
	"bpxe.org/pkg/flow_node/activity/task"
	"bpxe.org/pkg/flow_node/event/catch"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/assert"

	_ "github.com/stretchr/testify/assert"
)

var testDoc bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/boundary_event.bpmn", testdata, &testDoc)
}

func TestInterruptingEvent(t *testing.T) {
	testBoundaryEvent(t, "sig1listener", func(visited map[string]bool) {
		assert.False(t, visited["uninterrupted"])
		assert.True(t, visited["interrupted"])
		assert.True(t, visited["end"])
	}, event.NewSignalEvent("sig1"))
}

func TestNonInterruptingEvent(t *testing.T) {
	testBoundaryEvent(t, "sig2listener", func(visited map[string]bool) {
		assert.False(t, visited["interrupted"])
		assert.True(t, visited["uninterrupted"])
		assert.True(t, visited["end"])
	}, event.NewSignalEvent("sig2"))
}

func testBoundaryEvent(t *testing.T, boundary string, test func(visited map[string]bool), events ...event.ProcessEvent) {
	processElement := (*testDoc.Processes())[0]
	proc := process.New(&processElement, &testDoc)
	ready := make(chan bool)

	// explicit tracer
	tracer := tracing.NewTracer(context.Background())
	// this gives us some room when instance starts up
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 32))

	if instance, err := proc.Instantiate(process.WithTracer(tracer)); err == nil {
		if node, found := testDoc.FindBy(bpmn.ExactId("task")); found {
			if taskNode, found := instance.FlowNodeMapping().
				ResolveElementToFlowNode(node.(bpmn.FlowNodeInterface)); found {
				harness := taskNode.(*activity.Harness)
				aTask := harness.Activity().(*task.Task)
				aTask.SetBody(func(task *task.Task, ctx context.Context) flow_node.Action {
					select {
					case <-ready:
						return flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&task.Wiring.Outgoing)}
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

		err := instance.StartAll(context.Background())
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}

		listening := false
		activeBoundary := false

		for {
			if listening && activeBoundary {
				break
			}
			trace := <-traces
			t.Logf("%#v", trace)
			switch trace := trace.(type) {
			case catch.ActiveListeningTrace:
				// Ensure the boundary event listener is actually listening
				if id, present := trace.Node.Id(); present {
					if *id == boundary {
						// it is indeed listening
						listening = true
					}

				}
			case activity.ActiveBoundaryTrace:
				if id, present := trace.Node.Id(); present && trace.Start {
					if *id == "task" {
						// task has reached its active boundary
						activeBoundary = true
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
