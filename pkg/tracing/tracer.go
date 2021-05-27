// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tracing

import (
	"context"
	"sync"
)

type subscription struct {
	channel chan Trace
	ok      chan bool
}

type unsubscription struct {
	channel chan Trace
	ok      chan bool
}

type tracer struct {
	traces         chan Trace
	subscription   chan subscription
	unsubscription chan unsubscription
	terminate      chan struct{}
	subscribers    []chan Trace
	senders        sync.WaitGroup
}

func NewTracer(ctx context.Context) Tracer {
	tracer := tracer{
		subscribers:    make([]chan Trace, 0),
		traces:         make(chan Trace),
		subscription:   make(chan subscription),
		unsubscription: make(chan unsubscription),
		terminate:      make(chan struct{}),
	}
	go tracer.runner(ctx)
	return &tracer
}

func (t *tracer) runner(ctx context.Context) {
	var termination sync.Once
	for {
		select {
		case subscription := <-t.subscription:
			t.subscribers = append(t.subscribers, subscription.channel)
			subscription.ok <- true
		case unsubscription := <-t.unsubscription:
			pos := -1
			for i := range t.subscribers {
				if t.subscribers[i] == unsubscription.channel {
					pos = i
					break
				}
			}
			if pos >= 0 {
				l := len(t.subscribers) - 1
				// remove subscriber by replacing it with the last one
				t.subscribers[pos] = t.subscribers[l]
				t.subscribers[l] = nil
				// and truncating the list of subscribers
				t.subscribers = t.subscribers[:l]
				// (as we don't care about the order)
				unsubscription.ok <- true
			}
		case trace := <-t.traces:
			for _, subscriber := range t.subscribers {
				subscriber <- trace
			}
		case <-ctx.Done():
			// Start a termination waiting routine (only once)
			termination.Do(func() {
				go func() {
					// Wait until all senders have terminated
					t.senders.Wait()
					// Send an internal termination message
					t.terminate <- struct{}{}
				}()
			})
			// Let tracer continue to work for now
		case <-t.terminate:
			for _, subscriber := range t.subscribers {
				close(subscriber)
			}
			return
		}
	}
}

func (t *tracer) Subscribe() chan Trace {
	return t.SubscribeChannel(make(chan Trace))
}

func (t *tracer) SubscribeChannel(channel chan Trace) chan Trace {
	okChan := make(chan bool)
	sub := subscription{channel: channel, ok: okChan}
	t.subscription <- sub
	<-okChan
	return channel
}

func (t *tracer) Unsubscribe(c chan Trace) {
	okChan := make(chan bool)
	unsub := unsubscription{channel: c, ok: okChan}
loop:
	for {
		select {
		case <-c:
			continue loop
		case t.unsubscription <- unsub:
			continue loop
		case <-okChan:
			return
		}
	}
}

func (t *tracer) Trace(trace Trace) {
	t.traces <- trace
}

func (t *tracer) RegisterSender() SenderHandle {
	t.senders.Add(1)
	return &t.senders
}
