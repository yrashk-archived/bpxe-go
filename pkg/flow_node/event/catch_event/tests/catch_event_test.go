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
	"testing"
	"time"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	ev "bpxe.org/pkg/flow_node/event/catch_event"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"
)

func TestSignalEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "signalCatch",
		nil, event.NewSignalEvent("global_sig1"))
}

func TestNoneEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "noneCatch",
		nil, event.MakeNoneEvent())
}

func TestMessageEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "messageCatch",
		nil, event.NewMessageEvent("msg", nil))
}

func TestMultipleEvent(t *testing.T) {
	// either
	testEvent(t, "testdata/intermediate_catch_event_multiple.bpmn", "multipleCatch",
		nil, event.NewMessageEvent("msg", nil))
	// or
	testEvent(t, "testdata/intermediate_catch_event_multiple.bpmn", "multipleCatch",
		nil, event.NewSignalEvent("global_sig1"))
}

func TestMultipleParallelEvent(t *testing.T) {
	// both
	testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch",
		nil, event.NewMessageEvent("msg", nil), event.NewSignalEvent("global_sig1"))
	// either
	ch := make(chan bool)
	go func() {
		testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch",
			nil, event.NewMessageEvent("msg", nil))
		ch <- true
	}()
	go func() {
		testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch",
			nil, event.NewSignalEvent("global_sig1"))
		ch <- true
	}()
	select {
	case <-ch:
		t.Fatal("should not succeed")
	case <-time.After(time.Millisecond * 500):
	}

}

type eventInstance struct {
	id string
}

type eventInstanceBuilder struct{}

func (e eventInstanceBuilder) NewEventInstance(def bpmn.EventDefinitionInterface) event.Instance {
	switch d := def.(type) {
	case *bpmn.TimerEventDefinition:
		id, _ := d.Id()
		return eventInstance{id: *id}
	case *bpmn.ConditionalEventDefinition:
		id, _ := d.Id()
		return eventInstance{id: *id}
	default:
		return event.NewInstance(d)
	}
}

func TestTimerEvent(t *testing.T) {
	i := eventInstance{id: "timer_1"}
	b := eventInstanceBuilder{}
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "timerCatch", &b, event.MakeTimerEvent(i))
}

func TestConditionalEvent(t *testing.T) {
	i := eventInstance{id: "conditional_1"}
	b := eventInstanceBuilder{}
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "conditionalCatch", &b, event.MakeTimerEvent(i))
}

func testEvent(t *testing.T, filename string, nodeId string, eventInstanceBuilder event.InstanceBuilder, events ...event.ProcessEvent) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile(filename)
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if eventInstanceBuilder != nil {
		proc.SetEventInstanceBuilder(eventInstanceBuilder)
	}
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.SubscribeChannel(make(chan tracing.Trace, 64))
		err := instance.Run()
		if err != nil {
			t.Fatalf("failed to run the instance: %s", err)
		}
		resultChan := make(chan bool)
		go func() {
			for {
				trace := <-traces
				switch trace := trace.(type) {
				case ev.ActiveListeningTrace:
					if id, present := trace.Node.Id(); present {
						if *id == nodeId {
							// listening
							resultChan <- true
							return
						}

					}
				case tracing.ErrorTrace:
					t.Errorf("%#v", trace)
					resultChan <- false
					return
				default:
					t.Logf("%#v", trace)
				}
			}
		}()

		assert.True(t, <-resultChan)

		go func() {
			defer instance.Tracer.Unsubscribe(traces)
			for {
				trace := <-traces
				switch trace := trace.(type) {
				case flow.FlowTrace:
					if id, present := trace.Source.Id(); present {
						if *id == nodeId {
							// success!
							resultChan <- true
							return
						}

					}
				case tracing.ErrorTrace:
					t.Errorf("%#v", trace)
					resultChan <- false
					return
				default:
					t.Logf("%#v", trace)
				}
			}
		}()

		for _, evt := range events {
			_, err = instance.ConsumeProcessEvent(evt)
			assert.Nil(t, err)
		}

		assert.True(t, <-resultChan)

	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}

}
