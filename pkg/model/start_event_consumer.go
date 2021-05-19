// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/process/instance"
	"bpxe.org/pkg/tracing"
)

type startEventConsumer struct {
	process                                *process.Process
	parallel                               bool
	eventInstances, originalEventInstances []event.Instance
	ctx                                    context.Context
	consumptionLock                        sync.Mutex
	tracer                                 *tracing.Tracer
	events                                 []event.Event
	element                                bpmn.FlowNodeInterface
}

func newStartEventConsumer(
	ctx context.Context,
	tracer *tracing.Tracer,
	process *process.Process,
	startEvent *bpmn.StartEvent, builder event.InstanceBuilder) *startEventConsumer {
	var evCap int
	if startEvent.ParallelMultiple() {
		evCap = len(startEvent.EventDefinitions())
	} else {
		evCap = 1
	}
	consumer := &startEventConsumer{
		ctx:      ctx,
		process:  process,
		parallel: startEvent.ParallelMultiple(),
		tracer:   tracer,
		events:   make([]event.Event, 0, evCap),
		element:  startEvent,
	}
	consumer.eventInstances = make([]event.Instance, len(startEvent.EventDefinitions()))
	for k := range startEvent.EventDefinitions() {
		consumer.eventInstances[k] = builder.NewEventInstance(startEvent.EventDefinitions()[k])
	}
	consumer.originalEventInstances = consumer.eventInstances
	return consumer
}

func (s *startEventConsumer) ConsumeEvent(ev event.Event) (result event.ConsumptionResult, err error) {
	s.consumptionLock.Lock()
	defer s.consumptionLock.Unlock()
	defer s.tracer.Trace(EventInstantiationAttemptedTrace{Event: ev, Element: s.element})

	for i := range s.eventInstances {
		if ev.MatchesEventInstance(s.eventInstances[i]) {
			s.events = append(s.events, ev)
			if !s.parallel {
				goto instantiate
			} else {
				s.eventInstances[i] = s.eventInstances[len(s.eventInstances)-1]
				s.eventInstances = s.eventInstances[0 : len(s.eventInstances)-1]
				if len(s.eventInstances) == 0 {
					s.eventInstances = s.originalEventInstances
					goto instantiate
				}
			}
			break
		}
	}
	result = event.Consumed
	return
instantiate:
	var inst *instance.Instance
	inst, err = s.process.Instantiate(
		instance.WithContext(s.ctx),
		instance.WithTracer(s.tracer),
	)
	if err != nil {
		result = event.ConsumptionError
		return
	}
	for _, ev := range s.events {
		result, err = inst.ConsumeEvent(ev)
		if err != nil {
			result = event.ConsumptionError
			return
		}
	}
	s.events = s.events[:0]
	return
}
