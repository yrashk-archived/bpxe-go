// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package tracing

type subscription struct {
	channel chan Trace
	ok      chan bool
}

type unsubscription struct {
	channel chan Trace
	ok      chan bool
}

type Tracer struct {
	traces         chan Trace
	subscription   chan subscription
	unsubscription chan unsubscription
	subscribers    []chan Trace
}

func NewTracer() *Tracer {
	tracer := Tracer{
		subscribers:    make([]chan Trace, 0),
		traces:         make(chan Trace),
		subscription:   make(chan subscription),
		unsubscription: make(chan unsubscription),
	}
	go tracer.runner()
	return &tracer
}

func (t *Tracer) runner() {
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
		}
	}
}

// Subscribe creates a new unbuffered channel and subscribes it to
// traces from the Tracer
//
// Note that this channel should be continuously read from until unsubscribed
// from, otherwise, the Tracer will block.
func (t *Tracer) Subscribe() chan Trace {
	return t.SubscribeChannel(make(chan Trace))
}

// SubscribeChannel subscribes a channel to traces from the Tracer
//
// Note that this channel should be continuously read from (modulo
// buffering), otherwise, the Tracer will block.
func (t *Tracer) SubscribeChannel(channel chan Trace) chan Trace {
	okChan := make(chan bool)
	sub := subscription{channel: channel, ok: okChan}
	t.subscription <- sub
	<-okChan
	return channel
}

func (t *Tracer) Unsubscribe(c chan Trace) {
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

func (t *Tracer) Trace(trace Trace) {
	t.traces <- trace
}
