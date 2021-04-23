package flow

import (
	"fmt"
	"sync"

	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

// Represents a flow
type Flow struct {
	current         flow_node.FlowNodeInterface
	index           *int
	tracer          *tracing.Tracer
	flowNodeMapping *flow_node.FlowNodeMapping
	flowWaitGroup   *sync.WaitGroup
}

// Creates a new flow from a flow node
//
// The flow does nothing until it is explicitly started.
func NewFlow(current flow_node.FlowNodeInterface, tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping, flowWaitGroup *sync.WaitGroup) *Flow {
	return &Flow{
		current:         current,
		tracer:          tracer,
		flowNodeMapping: flowNodeMapping,
		flowWaitGroup:   flowWaitGroup,
	}
}

func (flow *Flow) handleSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow) {
	target, err := sequenceFlow.Target()
	if err == nil {
		if flowNode, found := flow.flowNodeMapping.ResolveElementToFlowNode(target); found {
			flow.current = flowNode
			var index int
			index, err = sequenceFlow.TargetIndex()
			if err != nil {
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				flow.index = nil
			}
			flow.index = new(int)
			*flow.index = index
		} else {
			flow.tracer.Trace(tracing.ErrorTrace{
				Error: errors.NotFoundError{Expected: fmt.Sprintf("flow node for element %#v", target)},
			})
		}
	} else {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
	}
}

func (flow *Flow) handleAdditionalSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow) {
	target, err := sequenceFlow.Target()
	if err == nil {
		if flowNode, found := flow.flowNodeMapping.ResolveElementToFlowNode(target); found {
			var index int
			newFlow := NewFlow(flowNode, flow.tracer, flow.flowNodeMapping, flow.flowWaitGroup)
			index, err = sequenceFlow.TargetIndex()
			if err != nil {
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				newFlow.index = nil
			}
			newFlow.index = new(int)
			*newFlow.index = index
			newFlow.Start()
		} else {
			flow.tracer.Trace(tracing.ErrorTrace{
				Error: errors.NotFoundError{Expected: fmt.Sprintf("flow node for element %#v", target)},
			})
		}
	} else {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
	}
}

// Starts the flow
func (flow *Flow) Start() {
	flow.flowWaitGroup.Add(1)
	go func() {
		defer flow.flowWaitGroup.Done()
		var action flow_node.Action
		for {
			if flow.index != nil {
				flow.current.Incoming(*flow.index)
			}
			action = flow.current.NextAction()
			switch a := action.(type) {
			case flow_node.FlowAction:
				sequenceFlows := a.SequenceFlows
				flow.tracer.Trace(tracing.FlowTrace{
					Source:        flow.current.Element(),
					SequenceFlows: sequenceFlows,
				})
				if len(a.SequenceFlows) > 0 {
					current := sequenceFlows[0]
					flow.handleSequenceFlow(current)

					rest := sequenceFlows[1:]
					for _, sequenceFlow := range rest {
						flow.handleAdditionalSequenceFlow(sequenceFlow)
					}
				} else {
					// nowhere to flow
					return
				}
			case flow_node.CompleteAction:
				flow.tracer.Trace(tracing.CompletionTrace{
					Node: flow.current.Element(),
				})
				return
			default:
			}
		}

	}()
}
