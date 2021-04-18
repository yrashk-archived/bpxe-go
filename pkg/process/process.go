package process

import (
	"bpxe.org/pkg/bpmn"
)

type Process struct {
	Element     *bpmn.Process
	Definitions *bpmn.Definitions
	instances   []*ProcessInstance
}

func MakeProcess(element *bpmn.Process, definitions *bpmn.Definitions) Process {
	return Process{
		Element:     element,
		Definitions: definitions,
		instances:   make([]*ProcessInstance, 0),
	}
}

func NewProcess(element *bpmn.Process, definitions *bpmn.Definitions) *Process {
	process := MakeProcess(element, definitions)
	return &process
}

func (process *Process) Instantiate() (instance *ProcessInstance, err error) {
	instance, err = NewProcessInstance(process)
	if err != nil {
		return
	}

	return
}
