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
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/model"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/require"
)

var testTimerStartEventInstantiation bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/instantiate_timer_start_event.bpmn", testdata, &testTimerStartEventInstantiation)
}

func TestTimerStartEventInstantiation(t *testing.T) {
	c := clock.NewMock()
	ctx := clock.ToContext(context.Background(), c)
	tracer := tracing.NewTracer(ctx)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	m := model.New(&testTimerStartEventInstantiation, model.WithContext(ctx), model.WithTracer(tracer))
	err := m.Run(ctx)
	require.Nil(t, err)
loop:
	for {
		select {
		case trace := <-traces:
			_, ok := trace.(flow.FlowTrace)
			// Should not flow
			require.False(t, ok)
		default:
			break loop
		}
	}
	// Advance clock by 1M
	c.Add(1 * time.Minute)
loop1:
	for {
		trace := <-traces
		switch trace := trace.(type) {
		case flow.VisitTrace:
			if idPtr, present := trace.Node.Id(); present {
				if *idPtr == "end" {
					// we've reached the desired outcome
					break loop1
				}
			}
		default:
			t.Logf("%#v", trace)
		}
	}
}

var testRecurringTimerStartEventInstantiation bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/instantiate_recurring_timer_start_event.bpmn", testdata, &testRecurringTimerStartEventInstantiation)
}

func TestRecurringTimerStartEventInstantiation(t *testing.T) {
	c := clock.NewMock()
	ctx := clock.ToContext(context.Background(), c)
	tracer := tracing.NewTracer(ctx)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	m := model.New(&testRecurringTimerStartEventInstantiation, model.WithContext(ctx), model.WithTracer(tracer))
	err := m.Run(ctx)
	require.Nil(t, err)
loop:
	for {
		select {
		case trace := <-traces:
			_, ok := trace.(flow.FlowTrace)
			// Should not flow
			require.False(t, ok)
		default:
			break loop
		}
	}
	// Test for some arbitrary number of recurrences (say, 10?)
	for i := 0; i < 10; i++ {
		// Advance clock by 1M
		c.Add(1 * time.Minute)
	loop1:
		for {
			trace := <-traces
			switch trace := trace.(type) {
			case flow.VisitTrace:
				if idPtr, present := trace.Node.Id(); present {
					if *idPtr == "end" {
						// we've reached the desired outcome
						break loop1
					}
				}
			default:
				t.Logf("%#v", trace)
			}
		}
	}
}
