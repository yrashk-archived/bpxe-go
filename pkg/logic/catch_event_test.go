// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package logic

import (
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"github.com/stretchr/testify/assert"
)

func TestCatchEventSatisfier_MatchSingle(t *testing.T) {
	catchEvent := bpmn.DefaultCatchEvent()

	sig1 := bpmn.DefaultSignalEventDefinition()
	sig1name := "sig1"
	sig1.SetSignalRef(&sig1name)

	catchEvent.SetSignalEventDefinitions([]bpmn.SignalEventDefinition{sig1})

	satisfier := NewCatchEventSatisfier(&catchEvent, event.DefaultInstanceBuilder{})

	var satisfied bool
	var chain int

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)
}

func TestCatchEventSatisfier_MatchMultiple(t *testing.T) {
	catchEvent := bpmn.DefaultCatchEvent()

	sig1 := bpmn.DefaultSignalEventDefinition()
	sig1name := "sig1"
	sig1.SetSignalRef(&sig1name)

	sig2 := bpmn.DefaultSignalEventDefinition()
	sig2name := "sig2"
	sig2.SetSignalRef(&sig2name)

	catchEvent.SetSignalEventDefinitions([]bpmn.SignalEventDefinition{sig1, sig2})

	satisfier := NewCatchEventSatisfier(&catchEvent, event.DefaultInstanceBuilder{})

	var satisfied bool
	var chain int

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)
	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig2name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)
}

func TestCatchEventSatisfier_MatchParallelMultiple(t *testing.T) {
	catchEvent := bpmn.DefaultCatchEvent()

	sig1 := bpmn.DefaultSignalEventDefinition()
	sig1name := "sig1"
	sig1.SetSignalRef(&sig1name)

	sig2 := bpmn.DefaultSignalEventDefinition()
	sig2name := "sig2"
	sig2.SetSignalRef(&sig2name)

	parallelMultiple := true
	catchEvent.SetParallelMultiple(&parallelMultiple)
	catchEvent.SetSignalEventDefinitions([]bpmn.SignalEventDefinition{sig1, sig2})

	satisfier := NewCatchEventSatisfier(&catchEvent, event.DefaultInstanceBuilder{})

	var satisfied bool
	var chain int

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.False(t, satisfied)
	assert.Equal(t, 0, chain)
	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig2name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	// Let's try this again, in a different order
	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig2name))
	assert.False(t, satisfied)
	assert.Equal(t, 0, chain)
	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

	// Now, let's supply two series of matching events but coming in partial
	// sequences

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.False(t, satisfied)
	assert.Equal(t, 0, chain)
	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.False(t, satisfied)
	assert.Equal(t, 1, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig2name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig2name))
	assert.True(t, satisfied)
	// the reason why chain here becomes 0 is that because that chain was satisfied
	// and removed, therefore this chain become indexed at 0
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

}

func TestCatchEventSatisfier_MatchParallelMultipleSingleEvent(t *testing.T) {
	catchEvent := bpmn.DefaultCatchEvent()

	sig1 := bpmn.DefaultSignalEventDefinition()
	sig1name := "sig1"
	sig1.SetSignalRef(&sig1name)

	parallelMultiple := true
	catchEvent.SetParallelMultiple(&parallelMultiple)
	catchEvent.SetSignalEventDefinitions([]bpmn.SignalEventDefinition{sig1})

	satisfier := NewCatchEventSatisfier(&catchEvent, event.DefaultInstanceBuilder{})

	var satisfied bool
	var chain int

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent(sig1name))
	assert.True(t, satisfied)
	assert.Equal(t, 0, chain)

	satisfied, chain = satisfier.Satisfy(event.NewSignalEvent("sig0"))
	assert.False(t, satisfied)
	assert.Equal(t, EventDidNotMatch, chain)
}
