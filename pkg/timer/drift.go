// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package timer

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
)

// DriftingClock is an augmentation of clock.Clock that allows
// to set/detect time drifts, both on wall and mock clocks
type DriftingClock struct {
	clock.Clock
	ch chan struct{}
}

// DriftOccurrence returns a channel that will signal every time
// there's either a backward drift or a forward drift higher than
// ForwardDriftTolerance
func (c *DriftingClock) DriftOccurrence() <-chan struct{} {
	return c.ch
}

const ForwardDriftTolerance = 3 * time.Second

func NewDriftingClock(ctx context.Context, c clock.Clock) *DriftingClock {
	dc := &DriftingClock{Clock: c, ch: make(chan struct{}, 1)}
	if _, ok := c.(*clock.Mock); !ok {
		// if it's a real clock, launch a monitor
		go func(ctx context.Context) {
			t := c.Now()
			for {
				select {
				case <-ctx.Done():
					return
				case <-c.After(time.Second * 1):
					if c.Now().Before(t) {
						// backward drift
						dc.DriftOccurred()
					} else if c.Now().Sub(t).Nanoseconds() > ForwardDriftTolerance.Nanoseconds() {
						// forward drift
						dc.DriftOccurrence()
					}
				}
			}
		}(ctx)
	}
	return dc
}

func (c *DriftingClock) DriftOccurred() {
	select {
	case c.ch <- struct{}{}:
	default:
	}
}

func (c *DriftingClock) Set(time time.Time) (err error) {
	if _, ok := c.Clock.(*clock.Mock); ok {
		mock := clock.NewMock()
		mock.Set(time)
		c.Clock = mock
		c.DriftOccurred()
	} else {
		err = fmt.Errorf("can't set time on non-mocked clock")
	}
	return
}

func (c *DriftingClock) Add(duration time.Duration) (err error) {
	if cl, ok := c.Clock.(*clock.Mock); ok {
		mock := clock.NewMock()
		mock.Set(cl.Now().Add(duration))
		c.Clock = mock
		c.DriftOccurred()
	} else {
		err = fmt.Errorf("can't set time on non-mocked clock")
	}
	return
}
