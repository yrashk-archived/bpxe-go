package flow

import (
	"fmt"
	"strings"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/expression"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

// Represents a flow
type Flow struct {
	id                id.Id
	definitions       *bpmn.Definitions
	current           flow_node.FlowNodeInterface
	index             *int
	tracer            *tracing.Tracer
	flowNodeMapping   *flow_node.FlowNodeMapping
	flowWaitGroup     *sync.WaitGroup
	idGenerator       id.Generator
	actionTransformer flow_node.ActionTransformer
	terminate         flow_node.Terminate
	sequenceFlowId    *string
}

func (flow *Flow) SequenceFlow() *sequence_flow.SequenceFlow {
	if flow.sequenceFlowId == nil {
		return nil
	} else {
		seqFlow, present := flow.definitions.FindBy(bpmn.ExactId(*flow.sequenceFlowId).
			And(bpmn.ElementType((*bpmn.SequenceFlow)(nil))))
		if present {
			return sequence_flow.NewSequenceFlow(seqFlow.(*bpmn.SequenceFlow), flow.definitions)
		} else {
			return nil
		}
	}
}

func (flow *Flow) Id() id.Id {
	return flow.id
}

func (flow *Flow) SetTerminate(terminate flow_node.Terminate) {
	flow.terminate = terminate
}

// Creates a new flow from a flow node
//
// The flow does nothing until it is explicitly started.
func NewFlow(definitions *bpmn.Definitions,
	current flow_node.FlowNodeInterface, tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping, flowWaitGroup *sync.WaitGroup,
	idGenerator id.Generator, actionTransformer flow_node.ActionTransformer) *Flow {
	return &Flow{
		id:                idGenerator.New(),
		definitions:       definitions,
		current:           current,
		tracer:            tracer,
		flowNodeMapping:   flowNodeMapping,
		flowWaitGroup:     flowWaitGroup,
		idGenerator:       idGenerator,
		actionTransformer: actionTransformer,
	}
}

func (flow *Flow) testSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow, unconditional bool) (result bool, err error) {
	if unconditional {
		result = true
		return
	}
	if expr, present := sequenceFlow.SequenceFlow.ConditionExpression(); present {
		switch e := expr.Expression.(type) {
		case *bpmn.FormalExpression:
			var lang string

			if language, present := e.Language(); present {
				lang = *language
			} else {
				lang = *flow.definitions.ExpressionLanguage()
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
		case *bpmn.Expression:
			// informal expression, can't execute
			result = true
		}
	} else {
		result = true
	}

	return
}

func (flow *Flow) handleSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow, unconditional bool,
	actionTransformer flow_node.ActionTransformer, terminate flow_node.Terminate) (flowed bool) {
	ok, err := flow.testSequenceFlow(sequenceFlow, unconditional)
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
			if idPtr, present := sequenceFlow.Id(); present {
				flow.sequenceFlowId = idPtr
			} else {
				flow.tracer.Trace(tracing.ErrorTrace{
					Error: errors.NotFoundError{Expected: fmt.Sprintf("id for sequence flow %#v", sequenceFlow)},
				})
			}
			flow.current = flowNode
			flow.terminate = terminate
			flow.tracer.Trace(VisitTrace{Node: flow.current.Element()})
			flow.actionTransformer = actionTransformer
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

func (flow *Flow) handleAdditionalSequenceFlow(sequenceFlow *sequence_flow.SequenceFlow, unconditional bool,
	actionTransformer flow_node.ActionTransformer, terminate flow_node.Terminate) (flowId id.Id, flowed bool) {
	ok, err := flow.testSequenceFlow(sequenceFlow, unconditional)
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
			newFlow := NewFlow(flow.definitions, flowNode, flow.tracer, flow.flowNodeMapping, flow.flowWaitGroup,
				flow.idGenerator, actionTransformer)
			flowId = newFlow.Id()
			if idPtr, present := sequenceFlow.Id(); present {
				newFlow.sequenceFlowId = idPtr
			} else {
				flow.tracer.Trace(tracing.ErrorTrace{
					Error: errors.NotFoundError{Expected: fmt.Sprintf("id for sequence flow %#v", sequenceFlow)},
				})
			}
			newFlow.terminate = terminate
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

func (flow *Flow) termination() chan bool {
	if flow.terminate == nil {
		return nil
	} else {
		return flow.terminate(flow.sequenceFlowId)
	}
}

// Starts the flow
func (flow *Flow) Start() {
	flow.flowWaitGroup.Add(1)
	go func() {
		flow.tracer.Trace(NewFlowTrace{FlowId: flow.id})
		defer flow.flowWaitGroup.Done()
		flow.tracer.Trace(VisitTrace{Node: flow.current.Element()})
		for {
			if flow.index != nil {
				flow.current.Incoming(*flow.index)
			}
		await:
			select {
			case terminate := <-flow.termination():
				if terminate {
					flow.tracer.Trace(FlowTerminationTrace{
						FlowId: flow.Id(),
						Source: flow.current.Element(),
					})
					return
				} else {
					goto await
				}
			case action := <-flow.current.NextAction(flow):
				if flow.actionTransformer != nil {
					action = flow.actionTransformer(flow.sequenceFlowId, action)
				}
				switch a := action.(type) {
				case flow_node.ProbeAction:
					results := make([]int, 0)
					for i, seqFlow := range a.SequenceFlows {
						if result, err := flow.testSequenceFlow(seqFlow, false); err == nil {
							if result {
								results = append(results, i)
							}
						} else {
							flow.tracer.Trace(tracing.ErrorTrace{Error: err})
						}
					}
					a.ProbeReport(results)
				case flow_node.FlowAction:
					sequenceFlows := a.SequenceFlows
					if len(a.SequenceFlows) > 0 {
						unconditional := make([]bool, len(a.SequenceFlows))
						for _, index := range a.UnconditionalFlows {
							unconditional[index] = true
						}
						source := flow.current.Element()

						current := sequenceFlows[0]
						effectiveFlows := make([]Snapshot, 0)

						flowed := flow.handleSequenceFlow(current, unconditional[0], a.ActionTransformer, a.Terminate)

						if flowed {
							effectiveFlows = append(effectiveFlows, Snapshot{sequenceFlow: current, flowId: flow.Id()})
						}

						rest := sequenceFlows[1:]
						for i, sequenceFlow := range rest {
							flowId, flowed := flow.handleAdditionalSequenceFlow(sequenceFlow, unconditional[i+1],
								a.ActionTransformer, a.Terminate)
							if flowed {
								effectiveFlows = append(effectiveFlows, Snapshot{sequenceFlow: sequenceFlow, flowId: flowId})
							}
						}

						if len(effectiveFlows) > 0 {
							flow.tracer.Trace(FlowTrace{
								Source: source,
								Flows:  effectiveFlows,
							})
						} else {
							// no flows to continue with, abort
							flow.tracer.Trace(FlowTerminationTrace{
								FlowId: flow.Id(),
								Source: source,
							})
							return
						}

					} else {
						// nowhere to flow, abort
						return
					}
				case flow_node.CompleteAction:
					flow.tracer.Trace(CompletionTrace{
						Node: flow.current.Element(),
					})
					flow.tracer.Trace(FlowTerminationTrace{
						FlowId: flow.Id(),
						Source: flow.current.Element(),
					})
					return
				case flow_node.NoAction:
					flow.tracer.Trace(FlowTerminationTrace{
						FlowId: flow.Id(),
						Source: flow.current.Element(),
					})
					return
				default:
				}
			}
		}

	}()
}
