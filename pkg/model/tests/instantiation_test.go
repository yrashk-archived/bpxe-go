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
	"bpxe.org/pkg/model"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/require"
)

var testStartEventInstantiation bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/instantiate_start_event.bpmn", testdata, &testStartEventInstantiation)
}

func TestStartEventInstantiation(t *testing.T) {
	ctx := context.Background()
	tracer := tracing.NewTracer(ctx)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	m := model.New(&testStartEventInstantiation, model.WithContext(ctx), model.WithTracer(tracer))
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
	// Test simple event instantiation
	_, err = m.ConsumeEvent(event.NewSignalEvent("sig1"))
	require.Nil(t, err)
loop1:
	for {
		trace := <-traces
		switch trace := trace.(type) {
		case flow.VisitTrace:
			if idPtr, present := trace.Node.Id(); present {
				if *idPtr == "sig1a" {
					// we've reached the desired outcome
					break loop1
				}
			}
		default:
			t.Logf("%#v", trace)
		}
	}
}

func TestMultipleStartEventInstantiation(t *testing.T) {
	ctx := context.Background()
	tracer := tracing.NewTracer(ctx)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	m := model.New(&testStartEventInstantiation, model.WithContext(ctx), model.WithTracer(tracer))
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
	// Test multiple event instantiation
	_, err = m.ConsumeEvent(event.NewSignalEvent("sig2"))
	require.Nil(t, err)
loop1:
	for {
		trace := <-traces
		switch trace := trace.(type) {
		case flow.VisitTrace:
			if idPtr, present := trace.Node.Id(); present {
				if *idPtr == "sig2_sig3a" {
					// we've reached the desired outcome
					break loop1
				}
			}
		default:
			t.Logf("%#v", trace)
		}
	}
}

func TestParallelMultipleStartEventInstantiation(t *testing.T) {
	ctx := context.Background()
	tracer := tracing.NewTracer(ctx)
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))
	m := model.New(&testStartEventInstantiation, model.WithContext(ctx), model.WithTracer(tracer))
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
	// Test parallel multiple event instantiation
	signalEvent := event.NewSignalEvent("sig2")
	_, err = m.ConsumeEvent(signalEvent)
	require.Nil(t, err)
	sig3sent := false
loop1:
	for {
		trace := <-traces
		switch trace := trace.(type) {
		case model.EventInstantiationAttemptedTrace:
			if !sig3sent && signalEvent == trace.Event {
				if idPtr, present := trace.Element.Id(); present && *idPtr == "ParallelMultipleStartEvent" {
					_, err = m.ConsumeEvent(event.NewSignalEvent("sig3"))
					require.Nil(t, err)
					sig3sent = true
				}
			}
		case flow.VisitTrace:
			if idPtr, present := trace.Node.Id(); present {
				if *idPtr == "sig2_and_sig3a" {
					require.True(t, sig3sent)
					// we've reached the desired outcome
					break loop1
				}
			}
		default:
			t.Logf("%#v", trace)

		}

	}
}
