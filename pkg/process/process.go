// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package process

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/id"
)

type Process struct {
	Element     *bpmn.Process
	Definitions *bpmn.Definitions
	id.GeneratorBuilder
	instances            []*Instance
	eventInstanceBuilder event.InstanceBuilder
}

func (process *Process) SetEventInstanceBuilder(eventInstanceBuilder event.InstanceBuilder) {
	process.eventInstanceBuilder = eventInstanceBuilder
}

func MakeProcess(element *bpmn.Process, definitions *bpmn.Definitions, idGeneratorBuilder id.GeneratorBuilder) Process {
	return Process{
		Element:          element,
		Definitions:      definitions,
		GeneratorBuilder: idGeneratorBuilder,
		instances:        make([]*Instance, 0),
	}
}

func NewProcess(element *bpmn.Process, definitions *bpmn.Definitions) *Process {
	return NewProcessWithIdGeneratorBuilder(element, definitions, id.DefaultIdGeneratorBuilder)
}

func NewProcessWithIdGeneratorBuilder(element *bpmn.Process, definitions *bpmn.Definitions,
	idGeneratorBuilder id.GeneratorBuilder) *Process {
	process := MakeProcess(element, definitions, idGeneratorBuilder)
	return &process
}

func (process *Process) Instantiate() (instance *Instance, err error) {
	instance, err = NewInstance(process)
	if err != nil {
		return
	}

	return
}

func (process *Process) NewEventInstance(def bpmn.EventDefinitionInterface) event.Instance {
	if process.eventInstanceBuilder != nil {
		return process.eventInstanceBuilder.NewEventInstance(def)
	} else {
		return event.NewInstance(def)
	}
}
