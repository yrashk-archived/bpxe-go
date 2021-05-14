// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/process"
)

type Model struct {
	Element   *bpmn.Definitions
	processes []process.Process
}

func New(element *bpmn.Definitions) *Model {
	return NewWithIdGenerator(element, id.DefaultIdGeneratorBuilder)
}

func NewWithIdGenerator(element *bpmn.Definitions, idGeneratorBuilder id.GeneratorBuilder) *Model {
	procs := element.Processes()
	processes := make([]process.Process, len(*procs))
	for i := range *procs {
		processes[i] = process.Make(&(*procs)[i], element, idGeneratorBuilder)
	}
	return &Model{
		Element:   element,
		processes: processes,
	}
}

func (model *Model) Run() {
}

func (model *Model) FindProcessBy(f func(*process.Process) bool) (result *process.Process, found bool) {
	for i := range model.processes {
		if f(&model.processes[i]) {
			result = &model.processes[i]
			found = true
			return
		}
	}
	return
}
