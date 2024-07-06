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

	"2006-01-02",
	"2006-01-02 MST",
	"2006-01-02 -0700",
	"2006-01-02 15:04",
	"2006-01-02 3:04PM",
	"2006-01-02 15:04 MST",
	"2006-01-02 3:04PM -0700",

	"02-01-2006",
	"02-01-2006 MST",
	"02-01-2006 -0700",
	"02-01-2006 15:04",
	"02-01-2006 3:04PM",
	"02-01-2006 15:04 MST",
	"02-01-2006 3:04PM -0700",
}

var timeFormats []string = []string{
	"15:04",
	"3:04PM",
}

var monthFormats []string = []string{
	"Jan",
	"January",
	"01",

	"01/2006",
	"01/06",
	"01 2006",
	"01 06",
	"2006 01",
	"Jan 2006",
	"January 2006",
	"01-2006",
	"2006-01",
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
			t = time.Date(
				time.Now().Year(), // Update year if missing
				t.Month(),
				t.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
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

func ParseYearAndMonth(input string) (time.Time, error) {
	for _, format := range monthFormats {
		t, err := time.Parse(format, input)
		if err != nil {
			continue // Format mismatch. Try the next one.
		}
		return t, nil
	}

	return time.Time{}, errors.New("'" + input + "' does not match any month/year format!")
}

// DurationToString because time's implementation is ugly.
func DurationToString(d time.Duration) string {
	var output string
	if d < 0 {
		output = "-"
	}
	parts := []string{}

	d = d.Abs()

	switch {
	case d.Hours() >= 24:
		days := d.Hours() / 24
		parts = append(parts, fmt.Sprintf("%dd", int(days)))
		d %= 24 * time.Hour
		if days >= 5 {
			break
		}
	case d.Hours() >= 1:
		parts = append(parts, fmt.Sprintf("%dh", int(d.Hours())))
		d %= time.Hour
		fallthrough
	case d.Minutes() >= 1 && len(parts) < 2:
		parts = append(parts, fmt.Sprintf("%dm", int(d.Minutes())))
		d %= time.Minute
		fallthrough
	case d.Seconds() >= 1 && len(parts) < 2:
		parts = append(parts, fmt.Sprintf("%ds", int(d.Seconds())))
		d %= time.Second
	}

	return output + strings.Join(parts, " ")
}
