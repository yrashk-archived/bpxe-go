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

	_ "github.com/stretchr/testify/assert"
)

var testDoc bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/start.bpmn", testdata, &testDoc)
}

func TestEndEvent(t *testing.T) {
	processElement := (*testDoc.Processes())[0]
	proc := process.New(&processElement, &testDoc)
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
