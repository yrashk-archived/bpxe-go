// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tests

import (
	"testing"

	"bpxe.org/internal"
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/require"
)

var testCondExpr bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/condexpr.bpmn", testdata, &testCondExpr)
}

func TestTrueFormalExpression(t *testing.T) {
	processElement := (*testCondExpr.Processes())[0]
	proc := process.New(&processElement, &testCondExpr)
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
			case flow.CompletionTrace:
				if id, present := trace.Node.Id(); present {
					if *id == "end" {
						// success!
						break loop
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

var testCondExprFalse bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/condexpr_false.bpmn", testdata, &testCondExprFalse)
}

func TestFalseFormalExpression(t *testing.T) {
	processElement := (*testCondExprFalse.Processes())[0]
	proc := process.New(&processElement, &testCondExprFalse)
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
			case flow.CompletionTrace:
				if id, present := trace.Node.Id(); present {
					if *id == "end" {
						t.Fatalf("end should not have been reached")
					}
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			case flow.CeaseFlowTrace:
				// success
				break loop
			default:
				t.Logf("%#v", trace)
			}
		}
		instance.Tracer.Unsubscribe(traces)
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}

var testCondDataObject bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/condexpr_dataobject.bpmn", testdata, &testCondDataObject)
}

func TestCondDataObject(t *testing.T) {
	test := func(cond, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			processElement := (*testCondDataObject.Processes())[0]
			proc := process.New(&processElement, &testCondDataObject)
			if instance, err := proc.Instantiate(); err == nil {
				traces := instance.Tracer.Subscribe()
				// Set all data objects to false by default, except for `cond`
				for _, k := range []string{"cond1o", "cond2o"} {
					itemAware, found := instance.FindItemAwareByName(k)
					require.True(t, found)
					itemAware.Put(k == cond)
				}
				err := instance.Run()
				if err != nil {
					t.Fatalf("failed to run the instance: %s", err)
				}
			loop:
				for {
					trace := <-traces
					switch trace := trace.(type) {
					case flow.VisitTrace:
						if id, present := trace.Node.Id(); present {
							if *id == expected {
								break loop
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
	}
	t.Run("cond1o", test("cond1o", "a1"))
	t.Run("cond2o", test("cond2o", "a2"))
}
