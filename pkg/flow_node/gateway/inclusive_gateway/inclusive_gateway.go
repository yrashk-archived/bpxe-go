package inclusive_gateway

import (
	"fmt"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/gateway"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/sequence_flow"
	"bpxe.org/pkg/tracing"
)

type NoEffectiveSequenceFlows struct {
	*bpmn.InclusiveGateway
}

func (e NoEffectiveSequenceFlows) Error() string {
	ownId := "<unnamed>"
	if ownIdPtr, present := e.InclusiveGateway.Id(); present {
		ownId = *ownIdPtr
	}
	return fmt.Sprintf("No effective sequence flows found in exclusive gateway `%v`", ownId)
}

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
	flow     flow_interface.T
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type probingReport struct {
	result []int
	flowId id.Id
}

func (m probingReport) message() {}

type flowSync struct {
	response chan flow_node.Action
	flow     flow_interface.T
}

type InclusiveGateway struct {
	flow_node.FlowNode
	element                 *bpmn.InclusiveGateway
	runnerChannel           chan message
	defaultSequenceFlow     *sequence_flow.SequenceFlow
	nonDefaultSequenceFlows []*sequence_flow.SequenceFlow
	probing                 *chan flow_node.Action
	activated               *flowSync
	awaiting                []id.Id
	sync                    []chan flow_node.Action
	*flowTracker
	synchronized bool
}

func NewInclusiveGateway(process *bpmn.Process,
	definitions *bpmn.Definitions,
	inclusiveGateway *bpmn.InclusiveGateway,
	eventIngress event.ProcessEventConsumer,
	eventEgress event.ProcessEventSource,
	tracer *tracing.Tracer,
	flowNodeMapping *flow_node.FlowNodeMapping,
	flowWaitGroup *sync.WaitGroup,
) (node *InclusiveGateway, err error) {
	flowNode, err := flow_node.NewFlowNode(process,
		definitions,
		&inclusiveGateway.FlowNode,
		eventIngress, eventEgress,
		tracer, flowNodeMapping,
		flowWaitGroup)
	if err != nil {
		return
	}

	var defaultSequenceFlow *sequence_flow.SequenceFlow

	if seqFlow, present := inclusiveGateway.Default(); present {
		if node, found := flowNode.Process.FindBy(bpmn.ExactId(*seqFlow).
			And(bpmn.ElementType((*bpmn.SequenceFlow)(nil)))); found {
			defaultSequenceFlow = new(sequence_flow.SequenceFlow)
			*defaultSequenceFlow = sequence_flow.MakeSequenceFlow(
				node.(*bpmn.SequenceFlow),
				definitions,
			)
		} else {
			err = errors.NotFoundError{
				Expected: fmt.Sprintf("default sequence flow with ID %s", *seqFlow),
			}
			return nil, err
		}
	}

	nonDefaultSequenceFlows := flow_node.AllSequenceFlows(&flowNode.Outgoing,
		func(sequenceFlow *sequence_flow.SequenceFlow) bool {
			if defaultSequenceFlow == nil {
				return false
			}
			return *sequenceFlow == *defaultSequenceFlow
		},
	)

	node = &InclusiveGateway{
		FlowNode:                *flowNode,
		element:                 inclusiveGateway,
		runnerChannel:           make(chan message),
		nonDefaultSequenceFlows: nonDefaultSequenceFlows,
		defaultSequenceFlow:     defaultSequenceFlow,
		flowTracker:             newFlowTracker(tracer),
	}
	go node.runner()
	return
}

func (node *InclusiveGateway) runner() {
	defer node.flowTracker.shutdown()
	activity := node.flowTracker.activity()
	for {
		select {
		case msg := <-node.runnerChannel:
			switch m := msg.(type) {
			case probingReport:
				response := node.probing
				if response == nil {
					// Reschedule, there's no next action yet
					go func() {
						node.runnerChannel <- m
					}()
					continue
				}
				node.probing = nil
				flow := make([]*sequence_flow.SequenceFlow, 0)
				for _, i := range m.result {
					flow = append(flow, node.nonDefaultSequenceFlows[i])
				}

				switch len(flow) {
				case 0:
					// no successful non-default sequence flows
					if node.defaultSequenceFlow == nil {
						// exception (Table 13.2)
						node.FlowNode.Tracer.Trace(tracing.ErrorTrace{
							Error: NoEffectiveSequenceFlows{
								InclusiveGateway: node.element,
							},
						})
					} else {
						gateway.DistributeFlows(node.sync, []*sequence_flow.SequenceFlow{node.defaultSequenceFlow})
					}
				default:
					gateway.DistributeFlows(node.sync, flow)
				}
				node.synchronized = false
				node.activated = nil
			case nextActionMessage:
				if node.synchronized {
					if m.flow.Id() == node.activated.flow.Id() {
						// Activating flow returned
						node.sync = append(node.sync, m.response)
						node.probing = &m.response
						// and now we wait until the probe has returned
						continue
					}
				}
				if node.activated == nil {
					// Haven't been activated yet
					node.activated = &flowSync{response: m.response, flow: m.flow}
					node.awaiting = node.flowTracker.activeFlowsInCohort(m.flow.Id())
					node.sync = make([]chan flow_node.Action, 0)
				} else {
					// Already activated
					for i, awaitingId := range node.awaiting {
						if awaitingId == m.flow.Id() {
							// Remove
							node.awaiting[i] = node.awaiting[len(node.awaiting)-1]
							node.awaiting = node.awaiting[:len(node.awaiting)-1]
							break
						}
					}
					node.sync = append(node.sync, m.response)
				}
				node.trySync()

			default:
			}
		case <-activity:
			if node.activated != nil {
				node.awaiting = node.flowTracker.activeFlowsInCohort(node.activated.flow.Id())
				node.trySync()
			}
		}
	}
}

func (node *InclusiveGateway) trySync() {
	if !node.synchronized && len(node.awaiting) == 0 {
		// We've got everybody
		anId := node.activated.flow.Id()
		// Probe outgoing sequence flow using the first flow
		node.activated.response <- flow_node.ProbeAction{
			SequenceFlows: node.nonDefaultSequenceFlows,
			ProbeReport: func(indices []int) {
				node.runnerChannel <- probingReport{
					result: indices,
					flowId: anId,
				}
			},
		}

		node.synchronized = true
	}
}

func (node *InclusiveGateway) NextAction(flow flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response, flow: flow}
	return response
}

func (node *InclusiveGateway) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *InclusiveGateway) Element() bpmn.FlowNodeInterface {
	return node.element
}
