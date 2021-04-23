package flow

import (
	"fmt"
	"strings"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/expression"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

// Represents a flow
type Flow struct {
	definitions     *bpmn.Definitions
	current         flow_node.FlowNodeInterface
	index           *int
	tracer          *tracing.Tracer
	flowNodeMapping *flow_node.FlowNodeMapping
	flowWaitGroup   *sync.WaitGroup
}

// Creates a new flow from a flow node
//
// The flow does nothing until it is explicitly started.
func NewFlow(definitions *bpmn.Definitions,
	current flow_node.FlowNodeInterface, tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping, flowWaitGroup *sync.WaitGroup) *Flow {
	return &Flow{
		definitions:     definitions,
		current:         current,
		tracer:          tracer,
		flowNodeMapping: flowNodeMapping,
		flowWaitGroup:   flowWaitGroup,
	}
}

func (flow *Flow) testSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow) (result bool, err error) {
	if expr, present := sequenceFlow.SequenceFlow.ConditionExpression(); present {
		switch e := expr.Expression.(type) {
		case *bpmn.Expression:
			// informal expression, can't execute
			result = true
		case *bpmn.FormalExpression:
			var lang string

			if language, present := e.Language(); present {
				lang = *language
			} else if language, present := flow.definitions.ExpressionLanguage(); present {
				lang = *language
			}

			engine := expression.GetEngine(lang)
			source := strings.Trim(*e.TextPayload(), " \n")
			var compiled expression.CompiledExpression
			compiled, err = engine.CompileExpression(source)
			if err != nil {
				result = false
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				return
			}
			var abstractResult expression.Result
			abstractResult, err = engine.EvaluateExpression(compiled, nil)
			if err != nil {
				result = false
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				return
			}
			switch actualResult := abstractResult.(type) {
			case bool:
				result = actualResult
			default:
				err = errors.InvalidArgumentError{
					Expected: fmt.Sprintf("boolean result in conditionExpression (%s)",
						source),
					Actual: actualResult,
				}
				result = false
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				return
			}
		}
	} else {
		result = true
	}

	return
}

func (flow *Flow) handleSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow) (flowed bool) {
	ok, err := flow.testSequenceFlow(sequenceFlow)
	if err != nil {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
		return
	}
	if !ok {
		return
	}

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
			flowed = true
		} else {
			flow.tracer.Trace(tracing.ErrorTrace{
				Error: errors.NotFoundError{Expected: fmt.Sprintf("flow node for element %#v", target)},
			})
		}
	} else {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
	}
	return
}

func (flow *Flow) handleAdditionalSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow) (flowed bool) {
	ok, err := flow.testSequenceFlow(sequenceFlow)
	if err != nil {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
		return
	}
	if !ok {
		return
	}
	target, err := sequenceFlow.Target()
	if err == nil {
		if flowNode, found := flow.flowNodeMapping.ResolveElementToFlowNode(target); found {
			var index int
			newFlow := NewFlow(flow.definitions, flowNode, flow.tracer, flow.flowNodeMapping, flow.flowWaitGroup)
			index, err = sequenceFlow.TargetIndex()
			if err != nil {
				flow.tracer.Trace(tracing.ErrorTrace{Error: err})
				newFlow.index = nil
			}
			newFlow.index = new(int)
			*newFlow.index = index
			newFlow.Start()
			flowed = true
		} else {
			flow.tracer.Trace(tracing.ErrorTrace{
				Error: errors.NotFoundError{Expected: fmt.Sprintf("flow node for element %#v", target)},
			})
		}
	} else {
		flow.tracer.Trace(tracing.ErrorTrace{Error: err})
	}
	return
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
				if len(a.SequenceFlows) > 0 {
					source := flow.current.Element()

					current := sequenceFlows[0]
					effectiveFlows := make([]*sequence_flow.SequenceFlow, 0)

					flowed := flow.handleSequenceFlow(current)

					if flowed {
						effectiveFlows = append(effectiveFlows, current)
					}

					rest := sequenceFlows[1:]
					for _, sequenceFlow := range rest {
						flowed = flow.handleAdditionalSequenceFlow(sequenceFlow)
						if flowed {
							effectiveFlows = append(effectiveFlows, sequenceFlow)
						}
					}

					if len(effectiveFlows) > 0 {
						flow.tracer.Trace(tracing.FlowTrace{
							Source:        source,
							SequenceFlows: effectiveFlows,
						})
					} else {
						// no flows to continue with, abort
						flow.tracer.Trace(tracing.FlowTerminationTrace{
							Source: source,
						})
						return
					}

				} else {
					// nowhere to flow, abort
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
