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
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/tracing"
)

type Model struct {
	Element              *bpmn.Definitions
	processes            []process.Process
	eventConsumersLock   sync.RWMutex
	eventConsumers       []event.Consumer
	idGeneratorBuilder   id.GeneratorBuilder
	eventInstanceBuilder event.InstanceBuilder
	tracer               *tracing.Tracer
}

type Option func(context.Context, *Model) context.Context

func WithIdGenerator(builder id.GeneratorBuilder) Option {
	return func(ctx context.Context, model *Model) context.Context {
		model.idGeneratorBuilder = builder
		return ctx
	}
}

func WithEventInstanceBuilder(builder event.InstanceBuilder) Option {
	return func(ctx context.Context, model *Model) context.Context {
		model.eventInstanceBuilder = builder
		return ctx
	}
}

// WithContext will pass a given context to a new model
// instead of implicitly generated one
func WithContext(newCtx context.Context) Option {
	return func(ctx context.Context, model *Model) context.Context {
		return newCtx
	}
}

// WithTracer overrides model's tracer
func WithTracer(tracer *tracing.Tracer) Option {
	return func(ctx context.Context, model *Model) context.Context {
		model.tracer = tracer
		return ctx
	}
}

func New(element *bpmn.Definitions, options ...Option) *Model {
	procs := element.Processes()
	model := &Model{
		Element: element,
	}

	ctx := context.Background()

	for _, option := range options {
		ctx = option(ctx, model)
	}

	if model.idGeneratorBuilder == nil {
		model.idGeneratorBuilder = id.DefaultIdGeneratorBuilder
	}

	if model.eventInstanceBuilder == nil {
		model.eventInstanceBuilder = event.DefaultInstanceBuilder{}
	}

	if model.tracer == nil {
		model.tracer = tracing.NewTracer(ctx)
	}

	model.processes = make([]process.Process, len(*procs))

	for i := range *procs {
		model.processes[i] = process.Make(&(*procs)[i], element,
			process.WithIdGenerator(model.idGeneratorBuilder),
			process.WithEventIngress(model), process.WithEventEgress(model),
			process.WithEventInstanceBuilder(model),
			process.WithContext(ctx),
			process.WithTracer(model.tracer),
		)
	}
	return model
}

func (model *Model) Run(ctx context.Context) (err error) {
	// Setup process instantiation
	for i := range *model.Element.Processes() {
		instantiatingFlowNodes := (*model.Element.Processes())[i].InstantiatingFlowNodes()
		for j := range instantiatingFlowNodes {
			flowNode := instantiatingFlowNodes[j]

			switch node := flowNode.(type) {
			case *bpmn.StartEvent:
				err = model.RegisterEventConsumer(newStartEventConsumer(ctx,
					model.tracer,
					&model.processes[i],
					node, model.eventInstanceBuilder))
				if err != nil {
					return
				}
			case *bpmn.EventBasedGateway:
			case *bpmn.ReceiveTask:
			}
		}
	}
	return
}

func (model *Model) FindProcessBy(f func(*process.Process) bool) (result *process.Process, found bool) {
	for i := range model.processes {
		if f(&model.processes[i]) {
			result = &model.processes[i]
			found = true
			return
		}
	}
	return
}

func (model *Model) ConsumeEvent(ev event.Event) (result event.ConsumptionResult, err error) {
	model.eventConsumersLock.RLock()
	// We're copying the list of consumers here to ensure that
	// new consumers can subscribe during event forwarding
	eventConsumers := model.eventConsumers
	model.eventConsumersLock.RUnlock()
	result, err = event.ForwardEvent(ev, &eventConsumers)
	return
}

func (model *Model) RegisterEventConsumer(ev event.Consumer) (err error) {
	model.eventConsumersLock.Lock()
	model.eventConsumers = append(model.eventConsumers, ev)
	model.eventConsumersLock.Unlock()
	return
}

func (model *Model) NewEventInstance(def bpmn.EventDefinitionInterface) event.Instance {
	if model.eventInstanceBuilder != nil {
		return model.eventInstanceBuilder.NewEventInstance(def)
	} else {
		return event.NewInstance(def)
	}
}
