package bpmn

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSample(t *testing.T) {
	var sampleDoc Definitions
	var err error
	sample, err := testdata.ReadFile("testdata/sample.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(sample, &sampleDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processes := sampleDoc.Processes()
	assert.Equal(t, 1, len(*processes))
}

func TestParseSampleNs(t *testing.T) {
	var sampleDoc Definitions
	var err error
	sampleNs, err := testdata.ReadFile("testdata/sample_ns.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(sampleNs, &sampleDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	processes := sampleDoc.Processes()
	assert.Equal(t, 1, len(*processes))
}
