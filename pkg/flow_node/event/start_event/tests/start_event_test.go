package tests

import (
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	_ "github.com/stretchr/testify/assert"
)

func TestStartEvent(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/start.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(src, &testDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processElement := (*testDoc.Processes())[0]
	proc := process.NewProcess(&processElement, &testDoc)
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
			case tracing.FlowTrace:
				if id, present := trace.Source.Id(); present {
					if *id == "start" {
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
	} else {
		t.Fatalf("failed to instantiate the process: %s", err)
	}
}
