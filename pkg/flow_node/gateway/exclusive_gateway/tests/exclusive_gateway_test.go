// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tests

import (
	"encoding/xml"
	"errors"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node/gateway/exclusive_gateway"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"
)

func TestExclusiveGateway(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/exclusive_gateway.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.FlowTrace:
				for _, f := range trace.Flows {
					t.Logf("%#v", f.SequenceFlow())
					if target, err := f.SequenceFlow().Target(); err == nil {
						if id, present := target.Id(); present {
							assert.NotEqual(t, "task1", *id)
							if *id == "task2" {
								// reached task2 as expected
								break loop
							}
						} else {
							t.Fatalf("can't find target's Id %#v", target)
						}

					} else {
						t.Fatalf("can't find sequence flow target: %#v", err)
					}
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

func TestExclusiveGatewayWithDefault(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/exclusive_gateway_default.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.FlowTrace:
				for _, f := range trace.Flows {
					t.Logf("%#v", f.SequenceFlow())
					if target, err := f.SequenceFlow().Target(); err == nil {
						if id, present := target.Id(); present {
							assert.NotEqual(t, "task1", *id)
							assert.NotEqual(t, "task2", *id)
							if *id == "default_task" {
								// reached default_task as expected
								break loop
							}
						} else {
							t.Fatalf("can't find target's Id %#v", target)
						}

					} else {
						t.Fatalf("can't find sequence flow target: %#v", err)
					}
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

func TestExclusiveGatewayWithNoDefault(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/exclusive_gateway_no_default.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.FlowTrace:
				for _, f := range trace.Flows {
					t.Logf("%#v", f.SequenceFlow())
					if target, err := f.SequenceFlow().Target(); err == nil {
						if id, present := target.Id(); present {
							assert.NotEqual(t, "task1", *id)
							assert.NotEqual(t, "task2", *id)
						} else {
							t.Fatalf("can't find target's Id %#v", target)
						}

					} else {
						t.Fatalf("can't find sequence flow target: %#v", err)
					}
				}
			case tracing.ErrorTrace:
				var target exclusive_gateway.NoEffectiveSequenceFlows
				if errors.As(trace.Error, &target) {
					// success
					break loop
				} else {
					t.Fatalf("%#v", trace)
				}
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

func TestExclusiveGatewayIncompleteJoin(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/exclusive_gateway_multiple_incoming.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.Run()
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

		assert.Equal(t, 2, reached["exclusive"])
		assert.Equal(t, 2, reached["task2"])
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}
