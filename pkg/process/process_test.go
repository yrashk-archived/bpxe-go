// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package process

import (
	"context"
	"testing"

	"bpxe.org/internal"
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

var defaultDefinitions = bpmn.DefaultDefinitions()

var sampleDoc bpmn.Definitions

func init() {
	internal.LoadTestFile("testdata/sample.bpmn", testdata, &sampleDoc)
}

func TestExplicitInstantiation(t *testing.T) {
	if proc, found := sampleDoc.FindBy(bpmn.ExactId("sample")); found {
		process := New(proc.(*bpmn.Process), &defaultDefinitions)
		instance, err := process.Instantiate()
		assert.Nil(t, err)
		assert.NotNil(t, instance)
	} else {
		t.Fatalf("Can't find process `sample`")
	}
}

func TestCancellation(t *testing.T) {
	if proc, found := sampleDoc.FindBy(bpmn.ExactId("sample")); found {
		process := New(proc.(*bpmn.Process), &defaultDefinitions)

		ctx, cancel := context.WithCancel(context.Background())

		tracer := tracing.NewTracer(ctx)
		traces := tracer.SubscribeChannel(make(chan tracing.Trace, 128))

		instance, err := process.Instantiate(WithContext(ctx), WithTracer(tracer))
		assert.Nil(t, err)
		assert.NotNil(t, instance)

		cancel()

		cancelledFlowNodes := make([]bpmn.FlowNodeInterface, 0)

		for trace := range traces {
			switch trace := trace.(type) {
			case flow_node.CancellationTrace:
				cancelledFlowNodes = append(cancelledFlowNodes, trace.Node)
			default:
			}
		}

		assert.NotEmpty(t, cancelledFlowNodes)
	} else {
		t.Fatalf("Can't find process `sample`")
	}
}
