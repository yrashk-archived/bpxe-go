// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tracing

// Trace is an interface for actual data traces
type Trace interface {
	TraceInterface()
}

// SenderHandle is an interface for registered senders
type SenderHandle interface {
	// Done indicates that the sender has terminated
	Done()
}

type Tracer interface {
	// Subscribe creates a new unbuffered channel and subscribes it to
	// traces from the Tracer
	//
	// Note that this channel should be continuously read from until unsubscribed
	// from, otherwise, the Tracer will block.
	Subscribe() chan Trace

	// SubscribeChannel subscribes a channel to traces from the Tracer
	//
	// Note that this channel should be continuously read from (modulo
	// buffering), otherwise, the Tracer will block.
	SubscribeChannel(channel chan Trace) chan Trace

	// Unsubscribe removes channel from subscription list
	Unsubscribe(c chan Trace)

	// Trace sends in a trace to a tracer
	Trace(trace Trace)

	// RegisterSender registers a sender for termination purposes
	//
	// Once Sender is being terminated, before closing subscription channels,
	// it'll wait until all senders call SenderHandle.Done
	RegisterSender() SenderHandle

	// Done returns a channel that is closed when the tracer is done and terminated
	Done() chan struct{}
}
