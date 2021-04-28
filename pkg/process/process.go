package process

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/id"
)

type Process struct {
	Element     *bpmn.Process
	Definitions *bpmn.Definitions
	id.IdGeneratorBuilder
	instances            []*ProcessInstance
	eventInstanceBuilder event.InstanceBuilder
}

func (process *Process) SetEventInstanceBuilder(eventInstanceBuilder event.InstanceBuilder) {
	process.eventInstanceBuilder = eventInstanceBuilder
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

func (process *Process) NewEventInstance(def bpmn.EventDefinitionInterface) event.Instance {
	if process.eventInstanceBuilder != nil {
		return process.eventInstanceBuilder.NewEventInstance(def)
	} else {
		return event.NewInstance(def)
	}
}
