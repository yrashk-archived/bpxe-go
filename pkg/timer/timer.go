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
	"bpxe.org/pkg/clock"
	"bpxe.org/pkg/errors"

	"github.com/qri-io/iso8601"
)

func New(ctx context.Context, clock clock.Clock, definition bpmn.TimerEventDefinition) (ch chan bpmn.TimerEventDefinition, err error) {
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
		if repeatingInterval.Interval.Start == nil {
			now := clock.Now()
			repeatingInterval.Interval.Start = &now
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

func recurringTimer(ctx context.Context, clock clock.Clock, interval iso8601.RepeatingInterval, f func()) {
	if interval.Interval.Start == nil {
		panic("shouldn't happen, has to be always set, explicitly or by timer.New")
	}
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

	repetitions := interval.Repititions

	var endTimer <-chan time.Time
	var timer <-chan time.Time

	t := *interval.Interval.Start

	for {
		if repetitions == 0 {
			return
		}

		timer = clock.Until(t.Add(interval.Interval.Duration.Duration))

		if interval.Interval.End != nil {
			endTimer = clock.Until(*interval.Interval.End)
		}

		select {
		case <-endTimer:
			return
		case <-ctx.Done():
			return
		case t = <-timer:
			if interval.Interval.End == nil || interval.Interval.End.After(clock.Now()) {
				f()
			}
		}

		if repetitions > 0 {
			repetitions--
		}
	}
}

func dateTimeTimer(ctx context.Context, clock clock.Clock, t time.Time, f func()) {
	for {
		timer := clock.Until(t)
		select {
		case <-ctx.Done():
			return
		case <-timer:
			f()
			return
		}
	}
}
