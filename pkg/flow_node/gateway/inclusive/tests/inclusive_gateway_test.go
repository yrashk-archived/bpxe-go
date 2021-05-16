// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tests

import (
	"errors"
	"testing"

	"bpxe.org/internal"
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node/gateway/inclusive"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"
)

var testInclusiveGateway bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/inclusive_gateway.bpmn", testdata, &testInclusiveGateway)
}

func TestInclusiveGateway(t *testing.T) {
	processElement := (*testInclusiveGateway.Processes())[0]
	proc := process.New(&processElement, &testInclusiveGateway)
	tracer := tracing.NewTracer()
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 32))
	if instance, err := proc.Instantiate(process.WithTracer(tracer)); err == nil {
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		endReached := 0
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.FlowTrace:
				for _, f := range trace.Flows {
					t.Logf("%#v", f.SequenceFlow())
					if target, err := f.SequenceFlow().Target(); err == nil {
						if id, present := target.Id(); present {
							assert.NotEqual(t, "a3", *id)
							if *id == "end" {
								// reached end
								endReached++
								continue
							}
						} else {
							t.Fatalf("can't find target's Id %#v", target)
						}

					} else {
						t.Fatalf("can't find sequence flow target: %#v", err)
					}
				}
			case flow.CeaseFlowTrace:
				// should only reach `end` once
				assert.Equal(t, 1, endReached)
				break loop
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

var testInclusiveGatewayDefault bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/inclusive_gateway_default.bpmn", testdata, &testInclusiveGatewayDefault)
}

func TestInclusiveGatewayDefault(t *testing.T) {
	processElement := (*testInclusiveGatewayDefault.Processes())[0]
	proc := process.New(&processElement, &testInclusiveGatewayDefault)
	tracer := tracing.NewTracer()
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 32))
	if instance, err := proc.Instantiate(process.WithTracer(tracer)); err == nil {
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		endReached := 0
	loop:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.FlowTrace:
				for _, f := range trace.Flows {
					t.Logf("%#v", f.SequenceFlow())
					if target, err := f.SequenceFlow().Target(); err == nil {
						if id, present := target.Id(); present {
							assert.NotEqual(t, "a1", *id)
							assert.NotEqual(t, "a2", *id)
							if *id == "end" {
								// reached end
								endReached++
								continue
							}
						} else {
							t.Fatalf("can't find target's Id %#v", target)
						}

					} else {
						t.Fatalf("can't find sequence flow target: %#v", err)
					}
				}
			case flow.CeaseFlowTrace:
				// should only reach `end` once
				assert.Equal(t, 1, endReached)
				break loop
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

var testInclusiveGatewayNoDefault bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/inclusive_gateway_no_default.bpmn", testdata, &testInclusiveGatewayNoDefault)
}

func TestInclusiveGatewayNoDefault(t *testing.T) {
	processElement := (*testInclusiveGatewayNoDefault.Processes())[0]
	proc := process.New(&processElement, &testInclusiveGatewayNoDefault)
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
			case tracing.ErrorTrace:
				var target inclusive.NoEffectiveSequenceFlows
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
