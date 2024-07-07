package ian

import (
	"fmt"
	"strings"
	"time"
)

func DisplayCalendar(yearMonth, today time.Time, sunday, showWeeks bool, format func(monthDay int, isToday bool) string) (output string) {
	var weekdayOffset time.Weekday = 1
	if sunday {
		weekdayOffset = 0
	}

	header := yearMonth.Format("January 2006")

	if showWeeks {
		output += "    "
	}

	width := 4 * 7 // 4 chars per day, 7 days in a week
	output += fmt.Sprintf("%[1]*s\n", -width, fmt.Sprintf("%[1]*s", (width+len(header))/2, header))

	daysInMonth := 32 - time.Date(yearMonth.Year(), yearMonth.Month(), 32, 0, 0, 0, 0, time.UTC).Day()
	firstWeekdayInMonth := time.Date(yearMonth.Year(), yearMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Weekday()

	if showWeeks {
		output += strings.Repeat(" ", 3)
	}

	for i := 0; i < 7; i++ {
		output += fmt.Sprintf(" \033[2m%s\033[0m", time.Weekday((int(weekdayOffset) + i) % 7).String()[:3])
	}
	output += "\n"

	displayWeek := func(week int) string {
    var format string
    if _, currentWeek := today.ISOWeek(); currentWeek == week {
      format = "\033[22;1;37m"
    }
		return fmt.Sprintf(" \033[2m"+format+"%2d\033[0m", week)
	}

	emptyDays := int(firstWeekdayInMonth - weekdayOffset)
  if !sunday && firstWeekdayInMonth == 0 {
    emptyDays = 6
  }
	if emptyDays > 0 {
		if showWeeks {
			_, week := time.Date(yearMonth.Year(), yearMonth.Month(), 1, 0, 0, 0, 0, time.UTC).ISOWeek()
			output += displayWeek(week)
		}
		output += strings.Repeat(" ", 4*emptyDays)
	}

	for monthDay := 1; monthDay <= daysInMonth; monthDay++ {
		weekday := time.Weekday((int(firstWeekdayInMonth) + monthDay - 1) % 7)

		if showWeeks && weekday == weekdayOffset {
			_, week := time.Date(yearMonth.Year(), yearMonth.Month(), monthDay, 0, 0, 0, 0, time.UTC).ISOWeek()
			output += displayWeek(week)
		}

		isToday := yearMonth.Year() == today.Year() && yearMonth.Month() == today.Month() && monthDay == today.Day()

		format := format(monthDay, isToday)
		output += fmt.Sprintf(" "+format+"%3s\033[0m", fmt.Sprint(monthDay))

		if weekday == (weekdayOffset+6)%7 {
			output += "\n"
		}
	}

	output += "\n"
	return
}
