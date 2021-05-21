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
	"time"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/errors"

	"github.com/qri-io/iso8601"
)

func New(ctx context.Context, clock *DriftingClock, definition bpmn.TimerEventDefinition) (ch chan bpmn.TimerEventDefinition, err error) {
	timeDate, timeDatePresent := definition.TimeDate()
	timeCycle, timeCyclePresent := definition.TimeCycle()
	timeDuration, timeDurationPresent := definition.TimeDuration()
	switch {
	case timeDatePresent && !timeCyclePresent && !timeDurationPresent:
		ch = make(chan bpmn.TimerEventDefinition)
		var t time.Time
		t, err = iso8601.ParseTime(*timeDate.Expression.TextPayload())
		if err != nil {
			return
		}
		go dateTimeTimer(ctx, clock, t, func() {
			ch <- definition
		})
	case !timeDatePresent && timeCyclePresent && !timeDurationPresent:
		ch = make(chan bpmn.TimerEventDefinition)
		var repeatingInterval iso8601.RepeatingInterval
		repeatingInterval, err = iso8601.ParseRepeatingInterval(*timeCycle.Expression.TextPayload())
		if err != nil {
			return
		}
		go recurringTimer(ctx, clock, repeatingInterval, func() {
			ch <- definition
		})
	case !timeDatePresent && !timeCyclePresent && timeDurationPresent:
		ch = make(chan bpmn.TimerEventDefinition)
		var duration iso8601.Duration
		duration, err = iso8601.ParseDuration(*timeDuration.Expression.TextPayload())
		if err != nil {
			return
		}
		go dateTimeTimer(ctx, clock, clock.Now().Add(duration.Duration), func() {
			ch <- definition
		})
	default:
		err = errors.InvalidArgumentError{
			Expected: "one and only one of timeDate, timeCycle or timeDuration must be defined",
			Actual:   definition,
		}
		return
	}
	return
}

func recurringTimer(ctx context.Context, clock *DriftingClock, interval iso8601.RepeatingInterval, f func()) {
	if interval.Interval.Start != nil {
		ch := make(chan struct{})
		go dateTimeTimer(ctx, clock, *interval.Interval.Start, func() {
			ch <- struct{}{}
		})
		select {
		case <-ctx.Done():
			return
		case <-ch:
			break
		}
	}

	t := clock.Now().Add(interval.Interval.Duration.Duration)

	repetitions := interval.Repititions

	for {
		if repetitions == 0 {
			return
		}

		timer := clock.After(t.Sub(clock.Now()))
		var endTimer <-chan time.Time
		if interval.Interval.End != nil {
			endTimer = clock.After(interval.Interval.End.Sub(clock.Now()))
		}

		// If it's already time
		if t.Sub(clock.Now()).Nanoseconds() <= 0 {
			if interval.Interval.End == nil || interval.Interval.End.After(clock.Now()) {
				f()
			}
		} else {
			select {
			case <-endTimer:
				return
			case <-ctx.Done():
				return
			case <-clock.DriftOccurrence():
				continue
			case <-timer:
				if interval.Interval.End == nil || interval.Interval.End.After(clock.Now()) {
					f()
				}
			}
		}
		if repetitions > 0 {
			repetitions--
		}
		t = clock.Now().Add(interval.Interval.Duration.Duration)
	}
}

func dateTimeTimer(ctx context.Context, clock *DriftingClock, t time.Time, f func()) {
	for {
		// If it's now, we're done
		if t.Sub(clock.Now()).Nanoseconds() <= 0 {
			f()
			return
		}
		timer := clock.After(t.Sub(clock.Now()))
		select {
		case <-ctx.Done():
			return
		case <-clock.DriftOccurrence():
			continue
		case now := <-timer:
			if now.Equal(t) || now.After(t) {
				f()
				return
			}
			continue
		}
	}
}
