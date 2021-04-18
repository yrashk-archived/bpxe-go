package model

import (
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/process"
	"github.com/stretchr/testify/assert"
)

func exactId(s string) func(p *process.Process) bool {
	return func(p *process.Process) bool {
		if id, present := p.Element.Id(); present {
			return *id == s
		} else {
			return false
		}
	}
}

func TestFindProcess(t *testing.T) {
	var sampleDoc bpmn.Definitions
	var err error
	sample, err := testdata.ReadFile("testdata/sample.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(sample, &sampleDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	model := NewModel(&sampleDoc)
	if proc, found := model.FindProcessBy(exactId("sample")); found {
		if id, present := proc.Element.Id(); present {
			assert.Equal(t, *id, "sample")
		} else {
			t.Fatalf("found a process but it has no Id")
		}
	} else {
		t.Fatalf("can't find process `sample`")
	}

	if _, found := model.FindProcessBy(exactId("none")); found {
		t.Fatalf("found a process by a non-existent id")
	}
}
