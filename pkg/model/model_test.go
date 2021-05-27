// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"testing"

	"bpxe.org/internal"
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

var sampleDoc bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/sample.bpmn", testdata, &sampleDoc)
}

func TestFindProcess(t *testing.T) {
	model := New(&sampleDoc)
	if proc, found := model.FindProcessBy(exactId("sample")); found {
		if id, present := proc.Element.Id(); present {
			assert.Equal(t, *id, "sample")
		} else {
			t.Fatalf("found a process but it has no FlowNodeId")
		}
	} else {
		t.Fatalf("can't find process `sample`")
	}

	if _, found := model.FindProcessBy(exactId("none")); found {
		t.Fatalf("found a process by a non-existent id")
	}
}
