# iso8601
--
    import "github.com/qri-io/iso8601"

Package iso8601 implements parts of International Standards Organization (ISO)
8601: Information interchange - Representation of dates and times this package
augments the go standard library time package with some of the additional
definitions provided by ISO 8601.

Many will be familiar with the Internet Engineering Task Force (IETF) RFC3339
timestamp format used in JSON & lots of other places, which looks like this:

    2019-04-23T11:50:41Z

this format is a "profile" if ISO 8601. Datestamp formats between the two are
(generally) the same. ISO 8601 is a larger spec that includes other definitions
like a string-based duration and repeating interval formats.

It's worth pointing out that these extra definitions are a little fuzzy. This
package takes some liberties that may not be appropriate in all circumstances
like defining one month as 30 days. If you're looking for a package with
horology-level accuracy, this may not be the right package

This package is a work-in-progress and the API is not yet considered stable

## Usage

```go
const (
	// OneDay is defined 24 hours
	OneDay = time.Hour * 24
	// OneWeek is defined as 7 days
	OneWeek = time.Hour * 24 * 7
	// OneMonth is defined as 30 days
	OneMonth = time.Hour * 24 * 7 * 30
	// OneYear is defined as 365 days
	OneYear = time.Hour * 24 * 7 * 365
)
```

#### func  ParseTime

```go
func ParseTime(s string) (time.Time, error)
```
ParseTime defers to time.Parse with RFC3339 datestamp format

#### type Duration

```go
type Duration struct {
	String   string
	Duration time.Duration
}
```

Duration is a string representation of a time interval. From wikipedia:

    Duration defines the amount of intervening time in a time interval and are
    represented by the format P[n]Y[n]M[n]DT[n]H[n]M[n]S or P[n]W.
    In these representations, the [n] is replaced by the value for each of the
    date and time elements that follow the [n]. Leading zeros are not
    required, but the maximum number of digits for each element should be
    agreed to by the communicating parties. The capital letters
    P, Y, M, W, D, T, H, M, and S are designators for each of the date and
    time elements and are not replaced.

    P is the duration designator (for period) placed at the start of the
      duration representation.
    Y is the year designator that follows the value for the number of years.
    M is the month designator that follows the value for the number of months.
    W is the week designator that follows the value for the number of weeks.
    D is the day designator that follows the value for the number of days.
    T is the time designator that precedes the time components of the
      representation.
        H is the hour designator that follows the value for the number
          of hours.
        M is the minute designator that follows the value for the number
          of minutes.
        S is the second designator that follows the value for the number
          of seconds.
    For example, "P3Y6M4DT12H30M5S" represents a duration of
    "three years, six months, four days, twelve hours, thirty minutes,
     and five seconds".
    https://en.wikipedia.org/wiki/ISO_8601#Usage

#### func  ParseDuration

```go
func ParseDuration(s string) (d Duration, err error)
```
ParseDuration interprets a string representation into a Duration

#### type Interval

```go
type Interval struct {
	Start    *time.Time
	End      *time.Time
	Duration Duration
}
```

Interval is the intervening time between two points. From wikipedia:

    The amount of intervening time is expressed by a duration.
    The two time points (start and end) are expressed by either a combined
    date and time representation or just a date representation:
      <start>/<end>
      <start>/<duration>
      <duration>/<end>
      <duration>
    https://en.wikipedia.org/wiki/ISO_8601#Time_intervals

#### func  ParseInterval

```go
func ParseInterval(s string) (i Interval, err error)
```
ParseInterval creates an interval from a string

#### type RepeatingInterval

```go
type RepeatingInterval struct {
	// Repititions specifies the number of iterations remaining for this interval
	// 0 indicates no repititions remain
	// -1 indicates unbounded ("infinite") repititions
	Repititions int
	// Interval is the time interval to repeat
	Interval Interval
}
```

RepeatingInterval specifies a recurring time interval. From wikipedia:

    Repeating intervals are formed by adding "R[n]/" to the beginning of an
    interval expression, where R is used as the letter itself and [n] is
    replaced by the number of repetitions
    Leaving out the value for [n] means an unbounded number of repetitions.
    If the interval specifies the start then this is the start of the
    repeating interval.
    If the interval specifies the end but not the start (form 3 above),
    then this is the end of the repeating interval.
    For example, to repeat the interval of "P1Y2M10DT2H30M" five times
    starting at "2008-03-01T13:00:00Z", use
    "R5/2008-03-01T13:00:00Z/P1Y2M10DT2H30M".
    https://en.wikipedia.org/wiki/ISO_8601#Usage

#### func  ParseRepeatingInterval

```go
func ParseRepeatingInterval(s string) (ri RepeatingInterval, err error)
```
ParseRepeatingInterval interprets a string into a RepeatingValue

#### func (RepeatingInterval) Next

```go
func (ri RepeatingInterval) Next() RepeatingInterval
```
Next returns the subsequent RepeatingInterval, possibly decrementing the number
of remaning repititions
