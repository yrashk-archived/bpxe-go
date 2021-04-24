package process

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/id"
)

type Process struct {
	Element     *bpmn.Process
	Definitions *bpmn.Definitions
	id.IdGeneratorBuilder
	instances []*ProcessInstance
}

func MakeProcess(element *bpmn.Process, definitions *bpmn.Definitions, idGeneratorBuilder id.IdGeneratorBuilder) Process {
	return Process{
		Element:            element,
		Definitions:        definitions,
		IdGeneratorBuilder: idGeneratorBuilder,
		instances:          make([]*ProcessInstance, 0),
	}
}

func NewProcess(element *bpmn.Process, definitions *bpmn.Definitions) *Process {
	return NewProcessWithIdGeneratorBuilder(element, definitions, id.DefaultIdGeneratorBuilder)
}

func NewProcessWithIdGeneratorBuilder(element *bpmn.Process, definitions *bpmn.Definitions,
	idGeneratorBuilder id.IdGeneratorBuilder) *Process {
	process := MakeProcess(element, definitions, idGeneratorBuilder)
	return &process
}

func (process *Process) Instantiate() (instance *ProcessInstance, err error) {
	instance, err = NewProcessInstance(process)
	if err != nil {
		return
	}

	return
}
