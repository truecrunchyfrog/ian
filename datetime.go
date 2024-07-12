package ian

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type TimeRange struct {
	From, To time.Time
}

func (tr *TimeRange) IsZero() bool {
  return tr.From.IsZero() && tr.To.IsZero()
}

const DefaultTimeLayout string = "_2 Jan 15:04 MST 2006"

var formats []string = []string{
	"2/1",
	"2/1 MST",
	"2/1 -0700",
	"2/1 15:04",
	"2/1 3:04PM",
	"2/1 15:04 MST",
	"2/1 3:04PM -0700",

	"2/1/2006",
	"2/1/2006 MST",
	"2/1/2006 -0700",
	"2/1/2006 15:04",
	"2/1/2006 3:04PM",
	"2/1/2006 15:04 MST",
	"2/1/2006 3:04PM -0700",

	"2006/1/2",
	"2006/1/2 MST",
	"2006/1/2 -0700",
	"2006/1/2 15:04",
	"2006/1/2 3:04PM",
	"2006/1/2 15:04 MST",
	"2006/1/2 3:04PM -0700",

	"1/2 2006",
	"1/2 2006 MST",
	"1/2 2006 -0700",
	"1/2 2006 15:04",
	"1/2 2006 3:04PM",
	"1/2 2006 15:04 MST",
	"1/2 2006 3:04PM -0700",

	"2 Jan",
	"2 Jan MST",
	"2 Jan -0700",
	"2 Jan 15:04",
	"2 Jan 3:04PM",
	"2 Jan 15:04 MST",
	"2 Jan 3:04PM -0700",

	"2 January",
	"2 January MST",
	"2 January -0700",
	"2 January 15:04",
	"2 January 3:04PM",
	"2 January 15:04 MST",
	"2 January 3:04PM -0700",

	"2 Jan 2006",
	"2 Jan 2006 MST",
	"2 Jan 2006 -0700",
	"2 Jan 2006 15:04",
	"2 Jan 2006 3:04PM",
	"2 Jan 2006 15:04 MST",
	"2 Jan 2006 3:04PM -0700",

	"Jan 2",
	"Jan 2 MST",
	"Jan 2 -0700",
	"Jan 2 15:04",
	"Jan 2 3:04PM",
	"Jan 2 15:04 MST",
	"Jan 2 3:04PM -0700",

	"January 2",
	"January 2 MST",
	"January 2 -0700",
	"January 2 15:04",
	"January 2 3:04PM",
	"January 2 15:04 MST",
	"January 2 3:04PM -0700",

	"Jan 2 2006",
	"Jan 2 2006 MST",
	"Jan 2 2006 -0700",
	"Jan 2 2006 15:04",
	"Jan 2 2006 3:04PM",
	"Jan 2 2006 15:04 MST",
	"Jan 2 2006 3:04PM -0700",

	"January 2 2006",
	"January 2 2006 MST",
	"January 2 2006 -0700",
	"January 2 2006 15:04",
	"January 2 2006 3:04PM",
	"January 2 2006 15:04 MST",
	"January 2 2006 3:04PM -0700",

	"2006-1-2",
	"2006-1-2 MST",
	"2006-1-2 -0700",
	"2006-1-2 15:04",
	"2006-1-2 3:04PM",
	"2006-1-2 15:04 MST",
	"2006-1-2 3:04PM -0700",

	"2-1-2006",
	"2-1-2006 MST",
	"2-1-2006 -0700",
	"2-1-2006 15:04",
	"2-1-2006 3:04PM",
	"2-1-2006 15:04 MST",
	"2-1-2006 3:04PM -0700",
}

var timeFormats []string = []string{
	"15",
	"15:04",
	"3:04PM",
}

// ParseDateTime parses a string against many different formats.
// If timezone is omitted, the local is assumed (from global variable `UseTimezone`).
// If year is omitted, the current one is used.
func ParseDateTime(input string, timeZone *time.Location) (time.Time, error) {
	for _, format := range formats {
		t, err := time.ParseInLocation(format, input, timeZone)
		if err != nil {
			continue // Format mismatch. Try the next one.
		}
		if t.Year() == 0 {
			t = t.AddDate(time.Now().In(timeZone).Year(), 0, 0) // Default year
		}
		return t, nil
	}

	return time.Time{}, errors.New("'" + input + "' does not match any date/time format!")
}

func ParseTimeOnly(input string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, input)
		if err != nil {
			continue // Format mismatch. Try the next one.
		}
		return t, nil
	}

	return time.Time{}, errors.New("'" + input + "' does not match any time format!")
}

// DurationToString because time's implementation is ugly.
func DurationToString(d time.Duration) string {
	var output string
	if d < 0 {
		output = "-"
	}
	parts := []string{}

	d = d.Abs()

	if d.Hours() >= 24 {
		days := d.Hours() / 24
		parts = append(parts, fmt.Sprintf("%dd", int(days)))
		d %= 24 * time.Hour
		if days >= 5 {
			goto end
		}
	}
	if d.Hours() >= 1 {
		parts = append(parts, fmt.Sprintf("%dh", int(d.Hours())))
		d %= time.Hour
	}
	if d.Minutes() >= 1 && len(parts) < 2 {
		parts = append(parts, fmt.Sprintf("%dm", int(d.Minutes())))
		d %= time.Minute
	}
	if d.Seconds() >= 1 && len(parts) < 2 {
		parts = append(parts, fmt.Sprintf("%ds", int(d.Seconds())))
		d %= time.Second
	}

end:
	return output + strings.Join(parts, " ")
}

func IsTimeWithinPeriod(t time.Time, period TimeRange) bool {
	if period.From.After(period.To) {
		panic("invalid timerange")
	}
	return t.Compare(period.From) != -1 && t.Compare(period.To) != 1
}

// IsPeriodConfinedToPeriod returns true if period1 start and end is within period2.
func IsPeriodConfinedToPeriod(period1, period2 TimeRange) bool {
	return IsTimeWithinPeriod(period1.From, period2) && IsTimeWithinPeriod(period1.To, period2)
}

// DoPeriodsMeet compares two periods and returns true if they collide at some point, otherwise false.
func DoPeriodsMeet(period1, period2 TimeRange) bool {
	return IsTimeWithinPeriod(period1.From, period2) ||
		IsTimeWithinPeriod(period1.To, period2) ||
		IsTimeWithinPeriod(period2.From, period1) ||
		IsTimeWithinPeriod(period2.To, period1)
}
