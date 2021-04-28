package tests

import (
	"context"
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow"
	ev "bpxe.org/pkg/flow_node/event/intermediate_catch_event"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

func TestEventBasedGateway(t *testing.T) {
	testEventBasedGateway(t, func(reached map[string]int) {
		assert.Equal(t, 1, reached["task1"])
		assert.Equal(t, 1, reached["end"])
	}, event.NewSignalEvent("Sig1"))
	testEventBasedGateway(t, func(reached map[string]int) {
		assert.Equal(t, 1, reached["task2"])
		assert.Equal(t, 1, reached["end"])
	}, event.NewMessageEvent("Msg1", nil))
}

func testEventBasedGateway(t *testing.T, test func(map[string]int), events ...event.ProcessEvent) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/event_based_gateway.bpmn")
	if err != nil {
		t.Errorf("Can't read file: %v", err)
		return
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Errorf("XML unmarshalling error: %v", err)
		return
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
	if instance, err := proc.Instantiate(); err == nil {
		traces := instance.Tracer.Subscribe()
		err := instance.Run()
		if err != nil {
			t.Errorf("failed to run the instance: %s", err)
			return
		}

		resultChan := make(chan string)

		ctx, cancel := context.WithCancel(context.Background())

		go func(ctx context.Context) {
			for {
				select {
				case trace := <-traces:
					switch trace := trace.(type) {
					case ev.ActiveListeningTrace:
						if id, present := trace.Node.Id(); present {
							resultChan <- *id
						}
					case tracing.ErrorTrace:
						t.Errorf("%#v", trace)
						return
					default:
						t.Logf("%#v", trace)
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx)

		// Wait until both events are ready to listen
		assert.Regexp(t, "(signalEvent|messageEvent)", <-resultChan)
		assert.Regexp(t, "(signalEvent|messageEvent)", <-resultChan)

		cancel()

		ch := make(chan map[string]int)
		go func() {
			reached := make(map[string]int)
			for {
				trace := <-traces
				switch trace := trace.(type) {
				case flow.VisitTrace:
					if id, present := trace.Node.Id(); present {
						if counter, ok := reached[*id]; ok {
							reached[*id] = counter + 1
						} else {
							reached[*id] = 1
						}
					} else {
						t.Errorf("can't find element with Id %#v", id)
						ch <- reached
						return
					}
				case flow.CeaseFlowTrace:
					ch <- reached
					return
				case tracing.ErrorTrace:
					t.Errorf("%#v", trace)
					ch <- reached
					return
				default:
					t.Logf("%#v", trace)
				}
			}
		}()

		for _, evt := range events {
			_, err := instance.ConsumeProcessEvent(evt)
			if err != nil {
				t.Error(err)
				return
			}
		}

		test(<-ch)

		instance.Tracer.Unsubscribe(traces)
	} else {
		t.Errorf("failed to instantiate the process: %s", err)
		return
	}
}
