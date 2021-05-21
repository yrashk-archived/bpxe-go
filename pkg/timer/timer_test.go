// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package timer

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"testing"

	"bpxe.org/pkg/bpmn"
	"github.com/benbjohnson/clock"
	"github.com/qri-io/iso8601"
	"github.com/stretchr/testify/require"
)

func TestTimeDate(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	iso := "2021-05-21T16:43:43+00:00"
	timestamp := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&timestamp)
	require.Nil(t, err)
	definition.SetTimeDate(&timestamp)
	timer, err := New(context.Background(), c, definition)
	require.Nil(t, err)
	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}
	time, err := iso8601.ParseTime(iso)
	require.Nil(t, err)
	err = c.Set(time)
	require.Nil(t, err)
	<-timer
}

func TestTimeDuration(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	iso := "PT30M"
	duration := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&duration)
	require.Nil(t, err)
	definition.SetTimeDuration(&duration)
	timer, err := New(context.Background(), c, definition)
	require.Nil(t, err)
	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}
	dur, err := iso8601.ParseDuration(iso)
	require.Nil(t, err)
	err = c.Add(dur.Duration)
	require.Nil(t, err)
	<-timer
}

func TestTimeCycle(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	iso := "R3/PT30M"
	cycle := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&cycle)
	require.Nil(t, err)
	definition.SetTimeCycle(&cycle)
	timer, err := New(context.Background(), c, definition)
	require.NotNil(t, timer)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	interval, err := iso8601.ParseRepeatingInterval(iso)
	require.Nil(t, err)

	for i := 0; i < interval.Repititions; i++ {
		err = c.Add(interval.Interval.Duration.Duration)
		require.Nil(t, err)

		<-timer

		select {
		case <-timer:
			require.FailNow(t, "shouldn't happen")
		default:
		}
	}

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

}

func TestTimeCycleNoRep(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	iso := "R0/PT30M"
	cycle := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&cycle)
	require.Nil(t, err)
	definition.SetTimeCycle(&cycle)
	timer, err := New(context.Background(), c, definition)
	require.NotNil(t, timer)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	interval, err := iso8601.ParseRepeatingInterval(iso)
	require.Nil(t, err)

	err = c.Add(interval.Interval.Duration.Duration)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

}

func TestTimeCycleStartDate(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	date := "2021-05-21T16:43:43+00:00"
	iso := fmt.Sprintf("R3/%s/PT30M", date)
	cycle := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&cycle)
	require.Nil(t, err)
	definition.SetTimeCycle(&cycle)
	timer, err := New(context.Background(), c, definition)
	require.NotNil(t, timer)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	interval, err := iso8601.ParseRepeatingInterval(iso)
	require.Nil(t, err)

	err = c.Add(interval.Interval.Duration.Duration)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	err = c.Set(*interval.Interval.Start)
	require.Nil(t, err)
	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	for i := 0; i < interval.Repititions; i++ {
		err = c.Add(interval.Interval.Duration.Duration)
		require.Nil(t, err)

		<-timer

		select {
		case <-timer:
			require.FailNow(t, "shouldn't happen")
		default:
		}
	}

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

}

func TestTimeCycleIndefinitely(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := NewDriftingClock(ctx, clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	iso := "R/PT30M"
	cycle := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&cycle)
	require.Nil(t, err)
	definition.SetTimeCycle(&cycle)
	timer, err := New(ctx, c, definition)
	require.NotNil(t, timer)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	interval, err := iso8601.ParseRepeatingInterval(iso)
	require.Nil(t, err)

	// Do some number of iterations to show that it repeats indefinitely
	for i := 0; i < 3; i++ {
		err = c.Add(interval.Interval.Duration.Duration)
		require.Nil(t, err)

		<-timer

		select {
		case <-timer:
			require.FailNow(t, "shouldn't happen")
		default:
		}
	}

	cancel()

}

func TestTimeCycleEndDate(t *testing.T) {
	c := NewDriftingClock(context.Background(), clock.NewMock())

	definition := bpmn.DefaultTimerEventDefinition()
	date := "2021-05-21T16:43:43+00:00"
	iso := fmt.Sprintf("R/PT30M/%s", date)
	cycle := bpmn.AnExpression{}
	err := xml.NewDecoder(bytes.NewBufferString(
		fmt.Sprintf(`<bpmn:expression>%s</bpmn:expression>`, iso),
	)).Decode(&cycle)
	require.Nil(t, err)
	definition.SetTimeCycle(&cycle)
	timer, err := New(context.Background(), c, definition)
	require.NotNil(t, timer)
	require.Nil(t, err)

	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

	interval, err := iso8601.ParseRepeatingInterval(iso)
	require.Nil(t, err)

	// Do some number of iterations to show that it repeats indefinitely
	for i := 0; i < 3; i++ {
		err = c.Add(interval.Interval.Duration.Duration)
		require.Nil(t, err)

		<-timer

		select {
		case <-timer:
			require.FailNow(t, "shouldn't happen")
		default:
		}
	}

	// Shift to the end
	err = c.Set(*interval.Interval.End)
	require.Nil(t, err)
	// Add a duration
	err = c.Add(interval.Interval.Duration.Duration)
	require.Nil(t, err)

	// No more repetitions
	select {
	case <-timer:
		require.FailNow(t, "shouldn't happen")
	default:
	}

}
