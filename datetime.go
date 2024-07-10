package ian

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

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

// TODO consider changing this whole -1 second from end thing to just not include the edge case,
// or will that complicate this further??

func IsTimeWithinPeriod(t, periodStart, periodEnd time.Time) bool {
  if periodStart.After(periodEnd) {
    panic("periodStart is after periodEnd")
  }
  //        not before start                not after end
	return t.Compare(periodStart) != -1 && t.Compare(periodEnd) != 1
}

// IsPeriodConfinedToPeriod returns true if the period start1-end1 is within start2-end2.
func IsPeriodConfinedToPeriod(start1, end1, start2, end2 time.Time) bool {
  return IsTimeWithinPeriod(start1, start2, end2) && IsTimeWithinPeriod(end1, start2, end2)
}

// DoPeriodsMeet compares two periods and returns true if they collide at some point, otherwise false.
// start1 must not come after end1, and start2 must not come after end2.
func DoPeriodsMeet(start1, end1, start2, end2 time.Time) bool {
	return IsTimeWithinPeriod(start1, start2, end2) ||
		IsTimeWithinPeriod(end1, start2, end2) ||
		IsTimeWithinPeriod(start2, start1, end1) ||
		IsTimeWithinPeriod(end2, start1, end1)
}
