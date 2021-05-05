package sequence_flow

import (
	"fmt"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
)

type SequenceFlow struct {
	*bpmn.SequenceFlow
	definitions *bpmn.Definitions
}

func MakeSequenceFlow(sequenceFlow *bpmn.SequenceFlow, definitions *bpmn.Definitions) SequenceFlow {
	return SequenceFlow{
		SequenceFlow: sequenceFlow,
		definitions:  definitions,
	}
}

func NewSequenceFlow(sequenceFlow *bpmn.SequenceFlow, definitions *bpmn.Definitions) *SequenceFlow {
	seqFlow := MakeSequenceFlow(sequenceFlow, definitions)
	return &seqFlow
}

func (sequenceFlow *SequenceFlow) resolveId(id *string) (result bpmn.FlowNodeInterface, err error) {
	ownId, present := sequenceFlow.SequenceFlow.Id()
	if !present {
		err = errors.InvalidStateError{
			Expected: "SequenceFlow to have an Id",
			Actual:   "Id is not present",
		}
		return
	}
	var process *bpmn.Process
	for i := range *sequenceFlow.definitions.Processes() {
		proc := &(*sequenceFlow.definitions.Processes())[i]
		sequenceFlows := proc.SequenceFlows()
		for j := range *sequenceFlows {
			if idPtr, present := (*sequenceFlows)[j].Id(); present {
				if *idPtr == *ownId {
					process = proc
				}
			}
		}
	}
	if process == nil {
		err = errors.NotFoundError{
			Expected: fmt.Sprintf("sequence flow with ID %s", *ownId),
		}
		return
	}
	if flowNode, found := process.FindBy(
		bpmn.ExactId(*id).
			And(bpmn.ElementInterface((*bpmn.FlowNodeInterface)(nil)))); found {
		result = flowNode.(bpmn.FlowNodeInterface)
	} else {
		err = errors.NotFoundError{Expected: fmt.Sprintf("flow node with ID %s", *id)}
	}
	return
}

func (sequenceFlow *SequenceFlow) Source() (bpmn.FlowNodeInterface, error) {
	return sequenceFlow.resolveId(sequenceFlow.SequenceFlow.SourceRef())
}

func (sequenceFlow *SequenceFlow) Target() (bpmn.FlowNodeInterface, error) {
	return sequenceFlow.resolveId(sequenceFlow.SequenceFlow.TargetRef())
}

func (sequenceFlow *SequenceFlow) TargetIndex() (index int, err error) {
	var target bpmn.FlowNodeInterface
	target, err = sequenceFlow.Target()
	if err != nil {
		return
	}
	// ownId is present since Target() already checked for this
	ownId, _ := sequenceFlow.SequenceFlow.Id()
	incomings := target.Incomings()
	for i := range *incomings {
		if (*incomings)[i] == *ownId {
			index = i
			return
		}
	}
	err = errors.NotFoundError{Expected: fmt.Sprintf("matching incoming for %s", *ownId)}
	return
}
