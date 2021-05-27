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
	"time"

	"bpxe.org/internal"
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/clock"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/flow_node/event/catch"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/process/instance"
	"bpxe.org/pkg/timer"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/require"
)

var timerDoc bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/intermediate_catch_event_timer.bpmn", testdata, &timerDoc)
}

func TestCatchEvent_Timer(t *testing.T) {
	processElement := (*timerDoc.Processes())[0]
	proc := process.New(&processElement, &timerDoc)
	fanOut := event.NewFanOut()
	c := clock.NewMock()
	ctx := clock.ToContext(context.Background(), c)
	tracer := tracing.NewTracer(ctx)
	eventInstanceBuilder := event.DefinitionInstanceBuildingChain(
		timer.EventDefinitionInstanceBuilder(ctx, fanOut, tracer),
	)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	if i, err := proc.Instantiate(
		instance.WithTracer(tracer),
		instance.WithEventDefinitionInstanceBuilder(eventInstanceBuilder),
		instance.WithEventEgress(fanOut),
		instance.WithEventIngress(fanOut),
	); err == nil {
		err := i.StartAll(ctx)
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		advancedTime := false
	loop:
		for {
			trace := tracing.Unwrap(<-traces)
			switch trace := trace.(type) {
			case catch.ActiveListeningTrace:
				c.Add(1 * time.Minute)
				advancedTime = true
			case flow.CompletionTrace:
				if id, present := trace.Node.Id(); present {
					if *id == "end" {
						require.True(t, advancedTime)
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
		i.Tracer.Unsubscribe(traces)
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}
