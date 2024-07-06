package ian

import (
	"fmt"
	"strings"
	"time"
)

func DisplayCalendar(yearMonth, today time.Time, firstWeekday time.Weekday, showWeeks bool, format func(monthDay int, isToday bool) string) (output string) {
	sundayDate := time.Unix(0, 0).AddDate(0, 0, -int(time.Thursday))

	header := yearMonth.Format("January 2006")
	width := 4 * 7 // 4 chars per day, 7 days in a week

	output += fmt.Sprintf("%[1]*s\n", -width, fmt.Sprintf("%[1]*s", (width+len(header))/2, header))

	daysInMonth := 32 - time.Date(yearMonth.Year(), yearMonth.Month(), 32, 0, 0, 0, 0, time.UTC).Day()
	firstWeekdayInMonth := time.Date(yearMonth.Year(), yearMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Weekday()

	for i := 0; i < 7; i++ {
		output += sundayDate.AddDate(0, 0, (int(firstWeekday)+i)%7).Format("Mon")
		if i != 7-1 {
			output += " "
		}
	}
	output += "\n"

	diff := func(x, y time.Weekday) time.Weekday {
		if x > y {
			return x - y
		}
		return y - x
	}

	output += strings.Repeat(" ", 4*int(diff(firstWeekday, firstWeekdayInMonth)))

	for monthDay := 1; monthDay <= daysInMonth; monthDay++ {
		weekday := time.Weekday((int(firstWeekdayInMonth) + monthDay) % 7)
    isToday := yearMonth.Year() == today.Year() && yearMonth.Month() == today.Month() && monthDay == today.Day()

		displaySlot := format(monthDay, isToday)
		output += fmt.Sprintf(displaySlot+"%3s\033[0m", displaySlot)

		if weekday == firstWeekday {
			output += "\n"
		} else {
			output += " "
		}
	}

	output += "\n"
	return
}
