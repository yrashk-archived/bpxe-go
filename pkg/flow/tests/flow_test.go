package tests

import (
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"

	_ "github.com/stretchr/testify/assert"
)

func TestTrueFormalExpression(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/condexpr.bpmn")
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
			case tracing.CompletionTrace:
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

func TestFalseFormalExpression(t *testing.T) {
	var testDoc bpmn.Definitions
	var err error
	src, err := testdata.ReadFile("testdata/condexpr_false.bpmn")
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
			case tracing.CompletionTrace:
				if id, present := trace.Node.Id(); present {
					if *id == "end" {
						t.Fatalf("end should not have been reached")
						break loop
					}
				}
			case tracing.ErrorTrace:
				t.Fatalf("%#v", trace)
			case tracing.CeaseFlowTrace:
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
