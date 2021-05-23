// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package clock

import (
	"sort"
	"sync"
	"time"
)

type after struct {
	time.Time
	ch chan time.Time
}

type afters []after

func (a afters) Len() int {
	return len(a)
}

func (a afters) Less(i, j int) bool {
	return a[i].Time.UnixNano() < a[j].Time.UnixNano()
}

func (a afters) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Mock clock is a fake clock that doesn't change
// unless explicitly instructed to.
type Mock struct {
	sync.RWMutex
	now     time.Time
	changes chan time.Time
	timers  afters
}

func (m *Mock) Now() time.Time {
	m.RLock()
	now := m.now
	m.RUnlock()
	return now
}

func (m *Mock) After(duration time.Duration) <-chan time.Time {
	m.Lock()
	ch := make(chan time.Time, 1)
	if duration.Nanoseconds() <= 0 {
		ch <- m.now
		close(ch)
		return ch
	}
	m.timers = append(m.timers, after{Time: m.now.Add(duration), ch: ch})
	m.Unlock()
	return ch
}

func (m *Mock) Until(t time.Time) <-chan time.Time {
	m.Lock()
	ch := make(chan time.Time, 1)
	if m.now.Equal(t) || m.now.After(t) {
		m.Unlock()
		ch <- m.now
		close(ch)
		return ch
	}
	m.timers = append(m.timers, after{Time: t, ch: ch})
	m.Unlock()
	return ch
}

func (m *Mock) Changes() <-chan time.Time {
	return m.changes
}

// NewMockAt creates a new Mock clock at a specific
// point in time
func NewMockAt(t time.Time) *Mock {
	source := &Mock{
		now:     t,
		changes: make(chan time.Time, 1),
	}
	return source
}

// NewMock creates a new Mock clock, set at the start of
// UNIX time
func NewMock() *Mock {
	return NewMockAt(time.Unix(0, 0))
}

func (m *Mock) Set(t time.Time) {
	m.Lock()
	m.lockedSet(t)
	m.Unlock()
}

func (m *Mock) Add(duration time.Duration) {
	m.Lock()
	m.lockedSet(m.now.Add(duration))
	m.Unlock()
}

// lockedSet should only be called when Mock is locked
func (m *Mock) lockedSet(t time.Time) {
	after := make([]after, 0, len(m.timers))
	sort.Sort(m.timers)
	for i := range m.timers {
		if m.timers[i].Equal(t) || m.timers[i].Before(t) {
			m.timers[i].ch <- t
		} else {
			after = append(after, m.timers[i])
		}
	}
	select {
	case m.changes <- t:
		// delivered changes notification
	default:
		// drop old time (we have a buffer of one)
		<-m.changes
		// push out new time
		m.changes <- t
	}
	m.timers = after
	m.now = t
}
