package ian

import (
	"errors"
	"fmt"
	"time"
)

type FrequencyType int

const (
	FreqSecondly FrequencyType = iota + 1
	FreqMinutely
	FreqHourly
	FreqDaily
	FreqWeekly
	FreqMonthly
	FreqYearly
)

// RecurrenceRule is an attempt to implement iCalendar's (RFC 5545) RRULE.
type RecurrenceRule struct {
	// Frequency is complemented by ByMonth and ByDay.
	Frequency FrequencyType
	// Interval is the interval for Frequency. An interval of 2 means every 2 Frequency.
	Interval uint

	BySecond   []uint8
	ByMinute   []uint8
	ByHour     []uint8
	ByDay      []time.Weekday
	ByMonthDay []uint8
	ByYearDay  []uint16
	ByWeekNo   []uint8
	ByMonth    []time.Month
	// BySetPos must be combined with another By* rule part.
	BySetPos []int16

	// WeekStart specifies the first day of the workweek.
	// Default is Monday = 1
	WeekStart time.Weekday

	// Count and Until are mutually exclusive. If both are left out (Count == 0 && Until.IsZero()), the recurrence is forever.

	// Count is the amount of times to repeat Frequency. The recurrence stops after that.
	Count uint
	// Until is the time when the recurrence should stop.
	Until time.Time
}

type RecurrenceRuleContext struct {
	Rule        RecurrenceRule
	Recurrences []time.Time
}

func (ctx *RecurrenceRuleContext) Next() error {
	if len(ctx.Recurrences) == 0 {
		return errors.New("empty recurrence set. there must be an initial datetime to base the recurrence on.")
	}

	if ctx.Rule.Count != 0 && len(ctx.Recurrences) >= int(ctx.Rule.Count) {
		return errors.New("recurrence amount passed 'Count'")
	}

	cursor := ctx.Recurrences[0]

	for !ctx.Recurrences[len(ctx.Recurrences)-1].Before(cursor) {
		// Next

    ctx.Rule.Frequency.IntoDeltaTime(int(ctx.Rule.Interval), cursor.Location())
	}

	return nil
}

func (freq FrequencyType) IntoDeltaTime(interval int, location *time.Location) (time.Time, error) {
	var seconds, minutes, hours, days, months, years int

	switch freq {
	case FreqSecondly:
    seconds = 1
  case FreqMinutely:
    minutes = 1
  case FreqHourly:
    hours = 1
  case FreqDaily:
    days = 1
  case FreqWeekly:
    days = 7
  case FreqMonthly:
    months = 1
  case FreqYearly:
    years = 1
  default:
    return time.Time{}, fmt.Errorf("invalid frequency type '%d'", freq)
	}

	return time.Date(
		years*interval,
		time.Month(months*interval),
		days*interval,
		hours*interval,
		minutes*interval,
		seconds*interval,
		0,
		location,
	), nil
}
