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

func NewModel(element *bpmn.Definitions) *Model {
	return NewModelWithIdGenerator(element, id.DefaultIdGeneratorBuilder)
}

func NewModelWithIdGenerator(element *bpmn.Definitions, idGeneratorBuilder id.GeneratorBuilder) *Model {
	procs := element.Processes()
	processes := make([]process.Process, len(*procs))
	for i := range *procs {
		processes[i] = process.MakeProcess(&(*procs)[i], element, idGeneratorBuilder)
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
