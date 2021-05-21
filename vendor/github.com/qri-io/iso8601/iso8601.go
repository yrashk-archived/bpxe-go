// Package iso8601 implements parts of International Standards Organization
// (ISO) 8601: Information interchange - Representation of dates and times
// this package augments the go standard library time package with some of the
// additional definitions provided by ISO 8601.
//
// Many will be familiar with the Internet Engineering Task Force (IETF) RFC3339
// timestamp format used in JSON & lots of other places, which looks like this:
//    2019-04-23T11:50:41Z
// this format is a "profile" if ISO 8601. Datestamp formats between the two are
// (generally) the same. ISO 8601 is a larger spec that includes other
// definitions like a string-based duration and repeating interval formats.
//
// It's worth pointing out that these extra definitions are a little fuzzy. This
// package takes some liberties that may not be appropriate in all circumstances
// like defining one month as 30 days. If you're looking for a package with
// horology-level accuracy, this may not be the right package
//
// This package is a work-in-progress and the API is not yet considered stable
package iso8601

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseTime defers to time.Parse with RFC3339 datestamp format
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// Duration is a string representation of a time interval. From wikipedia:
//    Duration defines the amount of intervening time in a time interval and are
//    represented by the format P[n]Y[n]M[n]DT[n]H[n]M[n]S or P[n]W.
//    In these representations, the [n] is replaced by the value for each of the
//    date and time elements that follow the [n]. Leading zeros are not
//    required, but the maximum number of digits for each element should be
//    agreed to by the communicating parties. The capital letters
//    P, Y, M, W, D, T, H, M, and S are designators for each of the date and
//    time elements and are not replaced.
//
//    P is the duration designator (for period) placed at the start of the
//      duration representation.
//    Y is the year designator that follows the value for the number of years.
//    M is the month designator that follows the value for the number of months.
//    W is the week designator that follows the value for the number of weeks.
//    D is the day designator that follows the value for the number of days.
//    T is the time designator that precedes the time components of the
//      representation.
//        H is the hour designator that follows the value for the number
//          of hours.
//        M is the minute designator that follows the value for the number
//          of minutes.
//        S is the second designator that follows the value for the number
//          of seconds.
//    For example, "P3Y6M4DT12H30M5S" represents a duration of
//    "three years, six months, four days, twelve hours, thirty minutes,
//     and five seconds".
//    https://en.wikipedia.org/wiki/ISO_8601#Usage
type Duration struct {
	duration string
	Duration time.Duration
}

// String returns Duration's string representation
func (d Duration) String() string {
	return d.duration
}

const (
	// OneDay is defined as 24 hours
	OneDay = time.Hour * 24
	// OneWeek is defined as 7 days
	OneWeek = time.Hour * 24 * 7
	// OneMonth is defined as 30 days
	OneMonth = time.Hour * 24 * 7 * 30
	// OneYear is defined as 365 days
	OneYear = time.Hour * 24 * 7 * 365
)

type unitType uint8

const (
	utNone unitType = iota
	utYear
	utMonth
	utWeek
	utDay
	utHour
	utMinute
	utSecond
)

func (u unitType) String() string {
	return map[unitType]string{
		utNone:   "none",
		utYear:   "year",
		utMonth:  "month",
		utWeek:   "week",
		utDay:    "day",
		utHour:   "hour",
		utMinute: "minute",
		utSecond: "second",
	}[u]
}

func (u unitType) Dur() time.Duration {
	return map[unitType]time.Duration{
		utNone:   time.Duration(0),
		utYear:   OneYear,
		utMonth:  OneMonth,
		utDay:    OneDay,
		utWeek:   OneWeek,
		utHour:   time.Hour,
		utMinute: time.Minute,
		utSecond: time.Second,
	}[u]
}

// ParseDuration interprets a string representation into a Duration
func ParseDuration(s string) (d Duration, err error) {
	d = Duration{duration: s}
	if len(s) < 3 {
		err = fmt.Errorf("string '%s' is too short", s)
		return
	}
	if s[0] != 'P' {
		err = fmt.Errorf("missing leading 'P' duration designator")
		return
	}

	digits := make([]rune, 0, 64)
	inTime := false
	durVal := 0
	prevUnit := utNone
	unit := utNone

	for _, r := range s[1:] {
		switch r {
		case 'T':
			inTime = true
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			digits = append(digits, r)
		default:
			durVal, err = strconv.Atoi(string(digits))
			if err != nil {
				return
			}
			digits = digits[:0]

			switch r {
			case 'Y':
				unit = utYear
			case 'M':
				if inTime {
					unit = utMinute
				} else {
					unit = utMonth
				}
			case 'W':
				// TODO (b5): spec says weeks can't be used in conjunction with other
				// duration units
				unit = utWeek
			case 'D':
				unit = utDay
			case 'H':
				unit = utHour
			case 'S':
				unit = utSecond
			default:
				err = fmt.Errorf("unrecognized duration character '%s'", string(r))
				return
			}

			if unit < prevUnit {
				err = fmt.Errorf("time units out of order: %s before %s", unit, prevUnit)
				return
			}
			d.Duration += time.Duration(durVal) * unit.Dur()
			prevUnit = unit
		}
	}

	return
}

// ParseInterval creates an interval from a string
func ParseInterval(s string) (i Interval, err error) {
	if len(s) < 3 {
		err = fmt.Errorf("string '%s' is too short", s)
		return
	}

	components := strings.Split(s, "/")
	switch len(components) {
	case 1:
		i.Duration, err = ParseDuration(components[0])
		return
	case 2:
		if len(components[0]) < 3 {
			err = fmt.Errorf("parsing start: string '%s' is too short", components[0])
			return
		} else if components[0][0] == 'P' {
			if i.Duration, err = ParseDuration(components[0]); err != nil {
				err = fmt.Errorf("parsing start duration: %s", err)
				return
			}
		} else {
			var t time.Time
			if t, err = ParseTime(components[0]); err != nil {
				err = fmt.Errorf("parsing start datestamp: %s", err)
				return
			}
			i.Start = &t
		}

		if len(components[1]) < 3 {
			err = fmt.Errorf("parsing end: string '%s' is too short", components[1])
			return
		} else if components[1][0] == 'P' {
			if i.Duration, err = ParseDuration(components[1]); err != nil {
				err = fmt.Errorf("parsing end duration: %s", err)
				return
			}
		} else {
			var t time.Time
			if t, err = ParseTime(components[1]); err != nil {
				err = fmt.Errorf("parsing end datestamp: %s", err)
				return
			}
			i.End = &t
		}

		if i.Start != nil && i.End != nil {
			i.Duration = Duration{
				// TODO (b5): implement time.Duration to 8601 Period String
				// String
				Duration: i.End.Sub(*i.Start),
			}
		}

	default:
		err = fmt.Errorf("too many interval designators (slashes)")
		return
	}
	return
}

// Interval is the intervening time between two points. From wikipedia:
//    The amount of intervening time is expressed by a duration.
//    The two time points (start and end) are expressed by either a combined
//    date and time representation or just a date representation:
//      <start>/<end>
//      <start>/<duration>
//      <duration>/<end>
//      <duration>
//    https://en.wikipedia.org/wiki/ISO_8601#Time_intervals
type Interval struct {
	Start    *time.Time
	End      *time.Time
	Duration Duration
}

// String returns Interval's string representation
func (i Interval) String() string {
	if i.Start != nil && i.End == nil {
		return fmt.Sprintf("%s/%s", i.Start.Format(time.RFC3339), i.Duration.String())
	} else if i.Start == nil && i.End != nil {
		return fmt.Sprintf("%s/%s", i.Duration.String(), i.End.Format(time.RFC3339))
	} else if i.Start != nil && i.End != nil {
		return fmt.Sprintf("%s/%s", i.Start.Format(time.RFC3339), i.End.Format(time.RFC3339))
	}
	return i.Duration.String()
}

// ParseRepeatingInterval interprets a string into a RepeatingValue
func ParseRepeatingInterval(s string) (ri RepeatingInterval, err error) {
	if len(s) < 3 {
		err = fmt.Errorf("string '%s' is too short", s)
		return
	}
	if s[0] != 'R' {
		err = fmt.Errorf("missing leading 'R' repeating designator")
		return
	}

	// default to infinite repititions
	ri.Repititions = -1
	digits := make([]rune, 0, 64)

RUNES:
	for i, r := range s[1:] {
		switch r {
		case '/':
			// TODO (b5): might be required
			// if len(s) < i+2 {
			// 	err = fmt.Errorf("missing interval value after interval designator (slash)")
			// 	return
			// }
			if ri.Interval, err = ParseInterval(s[i+2:]); err != nil {
				err = fmt.Errorf("parsing interval: %s", err)
				return
			}
			break RUNES
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			digits = append(digits, r)
		default:
			err = fmt.Errorf("unrecognized repeating interval character '%s'", string(r))
			return
		}
	}

	if len(digits) > 0 {
		if ri.Repititions, err = strconv.Atoi(string(digits)); err != nil {
			return
		}
	}

	return
}

// RepeatingInterval specifies a recurring time interval. From wikipedia:
//    Repeating intervals are formed by adding "R[n]/" to the beginning of an
//    interval expression, where R is used as the letter itself and [n] is
//    replaced by the number of repetitions
//    Leaving out the value for [n] means an unbounded number of repetitions.
//    If the interval specifies the start then this is the start of the
//    repeating interval.
//    If the interval specifies the end but not the start (form 3 above),
//    then this is the end of the repeating interval.
//    For example, to repeat the interval of "P1Y2M10DT2H30M" five times
//    starting at "2008-03-01T13:00:00Z", use
//    "R5/2008-03-01T13:00:00Z/P1Y2M10DT2H30M".
//    https://en.wikipedia.org/wiki/ISO_8601#Usage
type RepeatingInterval struct {
	// Repititions specifies the number of iterations remaining for this interval
	// 0 indicates no repititions remain
	// -1 indicates unbounded ("infinite") repititions
	Repititions int
	// Interval is the time interval to repeat
	Interval Interval
}

// String formats a repeatingInterval as a string value
func (ri RepeatingInterval) String() string {
	if ri.Repititions > 0 {
		return fmt.Sprintf("R%d/%s", ri.Repititions, ri.Interval)
	}
	return fmt.Sprintf("R/%s", ri.Interval)
}

// After returns the next instant an interval will occur from a given point
// in time.
// If the given time falls outside of the range specified by the interval,
// or no repitions of the interval remain, After returns the zero time instant
func (ri RepeatingInterval) After(t time.Time) time.Time {
	if ri.Repititions == 0 ||
		(ri.Interval.Start != nil && ri.Interval.Start.After(t)) ||
		(ri.Interval.End != nil && t.After(*ri.Interval.End)) {
		return time.Time{}
	}

	return t.Add(ri.Interval.Duration.Duration)
}

// NextRep returns the subsequent RepeatingInterval repitition,
// possibly decrementing the number of remaning repititions
func (ri RepeatingInterval) NextRep() RepeatingInterval {
	if ri.Repititions <= 0 {
		return ri
	}

	return RepeatingInterval{
		Repititions: ri.Repititions - 1,
		Interval:    ri.Interval,
	}
}
