package bpmn

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindByInArray(t *testing.T) {
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
	if e, found := sampleDoc.FindBy(ExactId("sample")); found {
		casted := e.(*Process)
		if casted != nil {
			id, present := casted.Id()
			assert.True(t, present)
			assert.NotNil(t, id)
			assert.Equal(t, *id, "sample")
		} else {
			t.Fatalf("found `sample` but it is not a `Process`")
		}
	} else {
		t.Fatalf("can't find process `sample`")
	}
}

func TestFindByInMaybe(t *testing.T) {
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
	if e, found := sampleDoc.FindBy(ExactId("x3cond")); found {
		casted := e.(*Expression)
		if casted != nil {
			id, present := casted.Id()
			assert.True(t, present)
			assert.NotNil(t, id)
			assert.Equal(t, *id, "x3cond")
		} else {
			t.Fatalf("found `x3cond` but it is not an `Expression`")
		}
	} else {
		t.Fatalf("can't find Expression `x3cond`")
	}
}

func TestFindByInSingleThroughInheritanceChain(t *testing.T) {
	var sampleDoc Definitions
	var err error
	sample, err := testdata.ReadFile("testdata/stdloop-example.bpmn")
	if err != nil {
		t.Fatalf("Can't read file: %v", err)
	}
	err = xml.Unmarshal(sample, &sampleDoc)
	if err != nil {
		t.Fatalf("XML unmarshalling error: %v", err)
	}
	if e, found := sampleDoc.FindBy(ExactId("stdloop")); found {
		casted := e.(*StandardLoopCharacteristics)
		if casted != nil {
			id, present := casted.Id()
			assert.True(t, present)
			assert.NotNil(t, id)
			assert.Equal(t, *id, "stdloop")
		} else {
			t.Fatalf("found `stdloop` but it is not an `StandardLoopCharacteristics`")
		}
	} else {
		t.Fatalf("can't find StandardLoopCharacteristics `stdloop`")
	}
}

func TestExactId(t *testing.T) {
	proc := DefaultProcess()
	s := "a"
	proc.SetId(&s)
	assert.True(t, ExactId("a")(&proc))
	assert.False(t, ExactId("b")(&proc))
}

func TestElementType(t *testing.T) {
	proc := DefaultProcess()
	assert.True(t, ElementType((*Process)(nil))(&proc))
	assert.False(t, ElementType((*Definitions)(nil))(&proc))
}

func TestElementInterface(t *testing.T) {
	proc := DefaultProcess()
	event := DefaultStartEvent()
	assert.True(t, ElementInterface((*FlowNodeInterface)(nil))(&event))
	assert.False(t, ElementInterface((*FlowNodeInterface)(nil))(&proc))
}

func TestElementPredicateAnd(t *testing.T) {
	proc := DefaultProcess()
	s := "a"
	proc.SetId(&s)
	assert.True(t, ElementType((*Process)(nil)).And(ExactId("a"))(&proc))
	assert.False(t, ElementType((*Process)(nil)).And(ExactId("b"))(&proc))
}

func TestElementPredicateOr(t *testing.T) {
	proc := DefaultProcess()
	s := "a"
	proc.SetId(&s)
	assert.True(t, ExactId("a").Or(ExactId("b"))(&proc))
	assert.False(t, ExactId("A").Or(ExactId("B"))(&proc))
}
