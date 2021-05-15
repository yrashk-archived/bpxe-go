// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package inclusive

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/flow"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/tracing"
)

type flowTracker struct {
	traces     <-chan tracing.Trace
	shutdownCh chan bool
	flows      map[id.Id]bpmn.Id
	activityCh chan struct{}
	lock       sync.RWMutex
	element    *bpmn.InclusiveGateway
}

func (tracker *flowTracker) activity() <-chan struct{} {
	return tracker.activityCh
}

func newFlowTracker(tracer *tracing.Tracer, element *bpmn.InclusiveGateway) *flowTracker {
	tracker := flowTracker{
		traces:     tracer.Subscribe(),
		shutdownCh: make(chan bool),
		flows:      make(map[id.Id]bpmn.Id),
		activityCh: make(chan struct{}),
		element:    element,
	}
	// Lock the tracker until it has caught up enough
	// to see the incoming flow for the node
	tracker.lock.Lock()
	go tracker.run()
	return &tracker
}

func (tracker *flowTracker) run() {
	// As per note in the constructor, we're starting in a locked mode
	locked := true
	// Flag for notifying the node about activity
	notify := false
	// Indicates whether the tracker has observed a flow
	// that reaches the node that uses this tracker.
	// This is important because if the node will invoke
	// `activeFlowsInCohort` before the tracker has caught up,
	// it'll return an empty list, and the node will assume that
	// there's no other flow to wait for, and will proceed (which
	// is incorrect)
	reachedNode := false
	for {
		select {
		case trace := <-tracker.traces:
			locked, notify, reachedNode = tracker.handleTrace(locked, trace, notify, reachedNode)
			// continue draining
			continue
		case <-tracker.shutdownCh:
			if locked {
				tracker.lock.Unlock()
			}
			return
		default:
			// Nothing else is coming in, unlock if locked
			if locked && reachedNode {
				tracker.lock.Unlock()
				if notify {
					tracker.activityCh <- struct{}{}
					notify = false
				}
				locked = false
			}
			// and now proceed with the second select to wait
			// for an event without doing busy work (this `default` clause)
		}
		select {
		case trace := <-tracker.traces:
			locked, notify, reachedNode = tracker.handleTrace(locked, trace, notify, reachedNode)
		case <-tracker.shutdownCh:
			if locked {
				tracker.lock.Unlock()
			}
			return
		}
	}
}

func (tracker *flowTracker) handleTrace(locked bool, trace tracing.Trace, notify bool, reachedNode bool) (bool, bool, bool) {
	if !locked {
		// Lock tracker records until messages are drained
		tracker.lock.Lock()
		locked = true
	}
	switch t := trace.(type) {
	case flow.FlowTrace:
		for _, snapshot := range t.Flows {
			// If we haven't reached the node
			if !reachedNode {
				// Try and see if this flow is the one that goes into it
				targetId := snapshot.SequenceFlow().TargetRef()
				if idPtr, present := tracker.element.Id(); present {
					reachedNode = *idPtr == *targetId
				}
			}
			if idPtr, present := t.Source.Id(); present {
				_, ok := tracker.flows[snapshot.Id()]
				_, isInclusive := t.Source.(*bpmn.InclusiveGateway)
				if !ok || isInclusive {
					tracker.flows[snapshot.Id()] = *idPtr
				}
			}
		}
		notify = true
	case flow.FlowTerminationTrace:
		delete(tracker.flows, t.FlowId)
		notify = true
	}
	return locked, notify, reachedNode
}

func (tracker *flowTracker) shutdown() {
	close(tracker.shutdownCh)
}

func (tracker *flowTracker) activeFlowsInCohort(flowId id.Id) (result []id.Id) {
	result = make([]id.Id, 0)
	tracker.lock.RLock()
	defer tracker.lock.RUnlock()
	if location, ok := tracker.flows[flowId]; ok {
		for k, v := range tracker.flows {
			if v == location && k != flowId {
				result = append(result, k)
			}
		}
	}
	return
}
