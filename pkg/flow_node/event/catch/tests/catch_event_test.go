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
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	ev "bpxe.org/pkg/flow_node/event/catch"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	"github.com/stretchr/testify/assert"
)

func TestSignalEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "signalCatch", nil, false, event.NewSignalEvent("global_sig1"))
}

func TestNoneEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "noneCatch", nil, false, event.MakeNoneEvent())
}

func TestMessageEvent(t *testing.T) {
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "messageCatch", nil, false, event.NewMessageEvent("msg", nil))
}

func TestMultipleEvent(t *testing.T) {
	// either
	testEvent(t, "testdata/intermediate_catch_event_multiple.bpmn", "multipleCatch", nil, false, event.NewMessageEvent("msg", nil))
	// or
	testEvent(t, "testdata/intermediate_catch_event_multiple.bpmn", "multipleCatch", nil, false, event.NewSignalEvent("global_sig1"))
}

func TestMultipleParallelEvent(t *testing.T) {
	// both
	testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch",
		nil, false, event.NewMessageEvent("msg", nil), event.NewSignalEvent("global_sig1"))
	// either
	testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch", nil, true, event.NewMessageEvent("msg", nil))
	testEvent(t, "testdata/intermediate_catch_event_multiple_parallel.bpmn", "multipleParallelCatch", nil, true, event.NewSignalEvent("global_sig1"))
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
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "timerCatch", &b, false, event.MakeTimerEvent(i))
}

func TestConditionalEvent(t *testing.T) {
	i := eventInstance{id: "conditional_1"}
	b := eventInstanceBuilder{}
	testEvent(t, "testdata/intermediate_catch_event.bpmn", "conditionalCatch", &b, false, event.MakeTimerEvent(i))
}

func testEvent(t *testing.T, filename string, nodeId string, eventInstanceBuilder event.InstanceBuilder, eventObservationOnly bool, events ...event.ProcessEvent) {
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
	proc := process.New(&processElement, &testDoc)
	if eventInstanceBuilder != nil {
		proc.SetEventInstanceBuilder(eventInstanceBuilder)
	}

	tracer := tracing.NewTracer(context.Background())
	traces := tracer.SubscribeChannel(make(chan tracing.Trace, 64))

	if instance, err := proc.Instantiate(process.WithTracer(tracer)); err == nil {
		err := instance.StartAll(context.Background())
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
			eventsToObserve := events
			for {
				trace := <-traces
				switch trace := trace.(type) {
				case ev.EventObservedTrace:
					if eventObservationOnly {
						for i := range eventsToObserve {
							if eventsToObserve[i] == trace.Event {
								eventsToObserve[i] = eventsToObserve[len(eventsToObserve)-1]
								eventsToObserve = eventsToObserve[:len(eventsToObserve)-1]
								break
							}
						}
						if len(eventsToObserve) == 0 {
							resultChan <- true
							return
						}
					}
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
