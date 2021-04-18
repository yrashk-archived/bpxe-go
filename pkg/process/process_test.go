package process

import (
	"encoding/xml"
	"testing"

	"bpxe.org/pkg/bpmn"
	"github.com/stretchr/testify/assert"
)

var defaultDefinitions = bpmn.DefaultDefinitions()

func TestExplicitInstantiation(t *testing.T) {
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

	if proc, found := sampleDoc.FindBy(bpmn.ExactId("sample")); found {
		process := NewProcess(proc.(*bpmn.Process), &defaultDefinitions)
		instance, err := process.Instantiate()
		assert.Nil(t, err)
		assert.NotNil(t, instance)
	} else {
		t.Fatalf("Can't find process `sample`")
	}
}
