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
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node/gateway/parallel"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"

	_ "bpxe.org/pkg/expression/expr"
)

var testParallelGateway bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/parallel_gateway_fork_join.bpmn", testdata, &testParallelGateway)
}

func TestParallelGateway(t *testing.T) {
	processElement := (*testParallelGateway.Processes())[0]
	proc := process.New(&processElement, &testParallelGateway)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.StartAll(context.Background())
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		reached := make(map[string]int)
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.VisitTrace:
				t.Logf("%#v", trace)
				if id, present := trace.Node.Id(); present {
					if counter, ok := reached[*id]; ok {
						reached[*id] = counter + 1
					} else {
						reached[*id] = 1
					}
				} else {
					t.Fatalf("can't find element with Id %#v", id)
				}
			case flow.CeaseFlowTrace:
				break loop
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)

		assert.Equal(t, 1, reached["task1"])
		assert.Equal(t, 1, reached["task2"])
		assert.Equal(t, 2, reached["join"])
		assert.Equal(t, 1, reached["end"])
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

var testParallelGatewayMtoN bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/parallel_gateway_m_n.bpmn", testdata, &testParallelGatewayMtoN)
}

func TestParallelGatewayMtoN(t *testing.T) {
	processElement := (*testParallelGatewayMtoN.Processes())[0]
	proc := process.New(&processElement, &testParallelGatewayMtoN)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.StartAll(context.Background())
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		reached := make(map[string]int)
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.VisitTrace:
				t.Logf("%#v", trace)
				if id, present := trace.Node.Id(); present {
					if counter, ok := reached[*id]; ok {
						reached[*id] = counter + 1
					} else {
						reached[*id] = 1
					}
				} else {
					t.Fatalf("can't find element with Id %#v", id)
				}
			case flow.CeaseFlowTrace:
				break loop
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)

		assert.Equal(t, 3, reached["joinAndFork"])
		assert.Equal(t, 1, reached["task1"])
		assert.Equal(t, 1, reached["task2"])
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

var testParallelGatewayNtoM bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/parallel_gateway_n_m.bpmn", testdata, &testParallelGatewayNtoM)
}

func TestParallelGatewayNtoM(t *testing.T) {
	processElement := (*testParallelGatewayNtoM.Processes())[0]
	proc := process.New(&processElement, &testParallelGatewayNtoM)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.StartAll(context.Background())
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		reached := make(map[string]int)
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.VisitTrace:
				if id, present := trace.Node.Id(); present {
					if counter, ok := reached[*id]; ok {
						reached[*id] = counter + 1
					} else {
						reached[*id] = 1
					}
				} else {
					t.Fatalf("can't find element with Id %#v", id)
				}
				t.Logf("%#v", reached)
			case flow.CeaseFlowTrace:
				break loop
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)

		assert.Equal(t, 2, reached["joinAndFork"])
		assert.Equal(t, 1, reached["task1"])
		assert.Equal(t, 1, reached["task2"])
		assert.Equal(t, 1, reached["task3"])
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

var testParallelGatewayIncompleteJoin bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/parallel_gateway_fork_incomplete_join.bpmn", testdata, &testParallelGatewayIncompleteJoin)
}

func TestParallelGatewayIncompleteJoin(t *testing.T) {
	processElement := (*testParallelGatewayIncompleteJoin.Processes())[0]
	proc := process.New(&processElement, &testParallelGatewayIncompleteJoin)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.StartAll(context.Background())
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		reached := make(map[string]int)
	loop:
		for trace := range traces {
			switch trace := trace.(type) {
			case parallel.IncomingFlowProcessedTrace:
				t.Logf("%#v", trace)
				if nodeIdPtr, present := trace.Node.Id(); present {
					if *nodeIdPtr == "join" {
						source, err := trace.Flow.SequenceFlow().Source()
						assert.Nil(t, err)
						if idPtr, present := source.Id(); present {
							if *idPtr == "task1" {
								// task1 already came in and has been
								// processed
								break loop
							}
						}
					}
				}
			case flow.FlowTrace:
				if idPtr, present := trace.Source.Id(); present {
					if *idPtr == "join" {
						t.Fatalf("should not flow from join")
					}
				}
			case flow.VisitTrace:
				t.Logf("%#v", trace)
				if id, present := trace.Node.Id(); present {
					if counter, ok := reached[*id]; ok {
						reached[*id] = counter + 1
					} else {
						reached[*id] = 1
					}
				} else {
					t.Fatalf("can't find element with Id %#v", id)
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)

		assert.Equal(t, 1, reached["task1"])
		assert.Equal(t, 0, reached["task2"])
		assert.Equal(t, 1, reached["join"])
		assert.Equal(t, 0, reached["end"])
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}
