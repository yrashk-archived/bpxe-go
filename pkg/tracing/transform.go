// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tracing

type Transform func(Trace) Trace

type traceTransformingTracer struct {
	tracer    Tracer
	transform Transform
}

func (t *traceTransformingTracer) Subscribe() chan Trace {
	return t.tracer.Subscribe()
}

func (t *traceTransformingTracer) SubscribeChannel(channel chan Trace) chan Trace {
	return t.tracer.SubscribeChannel(channel)
}

func (t *traceTransformingTracer) Unsubscribe(c chan Trace) {
	t.tracer.Unsubscribe(c)
}

func (t *traceTransformingTracer) Trace(trace Trace) {
	t.tracer.Trace(t.transform(trace))
}

func (t *traceTransformingTracer) RegisterSender() SenderHandle {
	return t.tracer.RegisterSender()
}

func NewTraceTransformingTracer(tracer Tracer, transform Transform) Tracer {
	return &traceTransformingTracer{tracer: tracer, transform: transform}
}
