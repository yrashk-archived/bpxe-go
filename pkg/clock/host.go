// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package clock

import (
	"context"
	"time"
)

type host struct {
	changes chan time.Time
}

func (h *host) Now() time.Time {
	return time.Now()
}

func (h *host) After(duration time.Duration) <-chan time.Time {
	expected := h.Now().Add(duration)
	in := time.After(duration)
	out := make(chan time.Time, 1)
	go func() {
		c := in
		for {
			t := <-c
			if t.Equal(expected) || t.After(expected) {
				out <- t
				return
			} else {
				c = time.After(expected.Sub(t))
			}
		}
	}()
	return out
}

func (h *host) Until(t time.Time) <-chan time.Time {
	return time.After(time.Until(t))
}

func (h *host) Changes() <-chan time.Time {
	return h.changes
}

const hostForwardDriftTolerance = 3 * time.Second

// Host is a clock source that uses time package as a source
// of time.
func Host(ctx context.Context) Clock {
	changes := make(chan time.Time)
	source := &host{changes: changes}
	go func(ctx context.Context) {
		t := source.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case t1 := <-source.After(time.Second * 1):
				if t1.Before(t1) {
					// backward drift
					source.changes <- t1
				} else if t1.Sub(t).Nanoseconds() > hostForwardDriftTolerance.Nanoseconds() {
					// forward drift
					source.changes <- t1
				}
			}
		}
	}(ctx)
	return source
}
