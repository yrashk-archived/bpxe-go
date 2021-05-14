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
}

func (tracker *flowTracker) activity() <-chan struct{} {
	return tracker.activityCh
}

func newFlowTracker(tracer *tracing.Tracer) *flowTracker {
	tracker := flowTracker{
		traces:     tracer.Subscribe(),
		shutdownCh: make(chan bool),
		flows:      make(map[id.Id]bpmn.Id),
		activityCh: make(chan struct{}),
	}
	go tracker.run()
	return &tracker
}

func (tracker *flowTracker) run() {
	for {
		select {
		case trace := <-tracker.traces:
			switch t := trace.(type) {
			case flow.FlowTrace:
				tracker.lock.Lock()
				for _, snapshot := range t.Flows {
					if idPtr, present := t.Source.Id(); present {
						_, ok := tracker.flows[snapshot.Id()]
						_, isInclusive := t.Source.(*bpmn.InclusiveGateway)
						if !ok || isInclusive {
							tracker.flows[snapshot.Id()] = *idPtr
						}
					}
				}
				tracker.lock.Unlock()
				tracker.activityCh <- struct{}{}
			case flow.FlowTerminationTrace:
				tracker.lock.Lock()
				delete(tracker.flows, t.FlowId)
				tracker.lock.Unlock()
				tracker.activityCh <- struct{}{}
			default:
				continue
			}
		case <-tracker.shutdownCh:
			return
		}
	}
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
