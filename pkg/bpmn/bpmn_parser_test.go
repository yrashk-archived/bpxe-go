// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

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
