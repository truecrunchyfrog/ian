package ian

import (
	"fmt"
	"image/color"
	"slices"
	"strings"
	"time"
)

func DisplayCalendar(
	location *time.Location,
	fromYear int, fromMonth time.Month,
	months int,
	firstWeekday time.Weekday,
	showWeeks bool,
	widthPerDay int,
	dayFmt func(y int, m time.Month, d int) (format string, fmtEntireSlot bool),
) (output string) {
	// Display the weekdays
	output += "  "
	for wd := 0; wd < 7; wd++ {
		weekday := time.Weekday((int(firstWeekday) + wd) % 7)

		dayString := weekday.String()
		if len(dayString) > widthPerDay {
			dayString = dayString[:widthPerDay]
		}
		output += fmt.Sprintf(" %"+fmt.Sprint(widthPerDay)+"s", dayString)
	}

	showWeekNumber := func(y int, m time.Month, d int) string {
		if showWeeks {
			_, week := time.Date(y, m, d, 0, 0, 0, 0, location).ISOWeek()
			return fmt.Sprint(week)
		}
		return "  "
	}

	// Display the month days, per month
	for unclampedMonth := fromMonth - 1; unclampedMonth < fromMonth+time.Month(months)-1; unclampedMonth++ {
		output += "\n"

		y := fromYear + int(unclampedMonth-1)/12
		m := unclampedMonth%12 + 1
		monthDays := 32 - time.Date(y, m, 32, 0, 0, 0, 0, location).Day()
		firstWeekdayInMonth := time.Date(y, m, 1, 0, 0, 0, 0, location).Weekday()

		var emptyDaysPadding string
		// Fill empty days until the first weekday
		for weekday := firstWeekday; weekday != firstWeekdayInMonth; weekday = (weekday + 1) % 7 {
			emptyDaysPadding += strings.Repeat(" ", widthPerDay+1)
		}

		output += emptyDaysPadding
		output += "   \033[1m" + m.String()[:3] + "\033[0m\n"

		if firstWeekdayInMonth != firstWeekday {
			output += showWeekNumber(y, m, 1)
			output += emptyDaysPadding
		}

		// Display the month's days.
		for d := 1; d <= monthDays; d++ {
			weekday := time.Weekday((int(firstWeekdayInMonth) + d - 1) % 7)

			if weekday == firstWeekday { // Week number/padding on week start
				output += showWeekNumber(y, m, d)
			}

			format, entireSlot := dayFmt(y, m, d)
			padding := strings.Repeat(" ", widthPerDay-2)
			if widthPerDay > 2 && entireSlot {
				format += padding
			} else {
				format = padding + format
			}
			output += fmt.Sprintf(" "+format+"%s\033[0m", fmt.Sprintf("%2d", d))

			if weekday == (firstWeekday+6)%7 { // Break line at end of week
				output += "\n"
			}
		}
	}

	return
}

func RgbToAnsiSeq(rgb color.RGBA, background bool) string {
	mode := 38
	if background {
		mode = 48
	}
	return fmt.Sprintf("\033[%d;2;%d;%d;%dm", mode, rgb.R, rgb.G, rgb.B)
}

// GetEventRgbAnsiSeq is a helper function to quickly get the color of an event based on its container.
func GetEventRgbAnsiSeq(event *Event, instance *Instance, background bool) string {
	calendar := event.Path.Calendar()
	var rgb color.RGBA
	if conf, err := instance.Config.GetContainerConfig(calendar); err == nil {
		rgb = conf.GetColor()
	} else {
		rgb = (&CalendarConfig{}).GetColor()
	}
	return RgbToAnsiSeq(rgb, background)
}

type eventEntry struct {
	event    *Event
	parent   *eventEntry
	children []*eventEntry
}

// possibleEntryDate returns a formatted date if it has not yet been shown (based on lastShownDate).
func possibleEntryDate(current time.Time, lastShownDate *time.Time) string {
	var date, year, month string

	if current.YearDay() != lastShownDate.YearDay() || current.Year() != lastShownDate.Year() {
		if current.Year() != lastShownDate.Year() {
			if !lastShownDate.IsZero() {
				year += fmt.Sprintf("%7s\n", "")
			}
			year += fmt.Sprintf("%-7s\n", current.Format("2006"))
		}
		if current.Month() != lastShownDate.Month() || current.Year() != lastShownDate.Year() {
			month = fmt.Sprintf("%7s\n", "")
		}
		date += current.Format("_2 Jan")
		*lastShownDate = current
	}

	now := time.Now().In(current.Location())
	var format string
	switch {
	case current.Year() < now.Year():
		fallthrough
	case current.YearDay() < now.YearDay() && current.Year() == now.Year():
		// Past days:
		format = "\033[3m"
		fallthrough
	case current.YearDay() != now.YearDay() || current.Year() != now.Year():
		// Not today:
		format = "\033[2m" + format
	case strings.TrimSpace(date) != "":
		// Today:
		format = "\033[1;30;47m"
	}
	return fmt.Sprintf("%s%s%s%-6s\033[0m ", year, month, format, date)
}

func displayEntry(instance *Instance, entry *eventEntry, showDates bool, lastShownDate *time.Time, location *time.Location) string {
	var output string

	var startFmt, endFmt string
	if entry.event != nil {
		// Itself

		prefix := "\033[1m"

		start := entry.event.Props.Start.In(location)
		end := entry.event.Props.End.In(location)
		if !entry.event.Props.IsAllDay() || entry.event.Props.Start.Location() != location {
			startFmt = start.Format("15")
			if start.Minute() != 0 {
				startFmt += start.Format(":04")
			}

			endFmt = end.Format("15")
			if end.Minute() != 0 {
				endFmt += end.Format(":04")
			}

			if len(entry.children) != 0 || start.Day() != end.Day() {
				prefix += startFmt
			} else {
				prefix += startFmt + " 🡲  " + endFmt
			}
		} else {
			prefix += "+"
		}

		prefix += "\033[22m "

		var suffix string
		if entry.event.Props.Recurrence.IsThereRecurrence() {
			suffix += " ⟳"
		}
		suffix += "\033[0m"

		pipes := displayPipes(instance, entry)

		if showDates {
			entryDate := possibleEntryDate(start, lastShownDate)
			entryDateLines := strings.Split(entryDate, "\n")
			for i := 0; i < len(entryDateLines); i++ {
				entryDateLines[i] += pipes
			}
			output += strings.Join(entryDateLines, "\n")
		} else {
      output += pipes
    }

		output += GetEventRgbAnsiSeq(entry.event, instance, false) + prefix + entry.event.Props.Summary + suffix + "\033[0m"
	}
	for _, child := range entry.children {
		// Children
		output += "\n" + displayEntry(instance, child, showDates, lastShownDate, location)
	}
	if entry.event != nil && (len(entry.children) != 0 || entry.event.Props.Start.In(location).Day() != entry.event.Props.End.In(location).Add(-time.Second).Day()) {
		// Tail

		pipes := displayPipes(instance, entry)
		innerPipes := displayPipes(instance, &eventEntry{parent: entry})

		output += "\n"

		if showDates {
			entryDate := possibleEntryDate(entry.event.Props.End.In(location), lastShownDate)
			entryDateLines := strings.Split(entryDate, "\n")
			for i := 0; i < len(entryDateLines); i++ {
				if i != len(entryDateLines)-1 {
					entryDateLines[i] += innerPipes
				} else {
					entryDateLines[i] += pipes
				}
			}
			output += strings.Join(entryDateLines, "\n")
		} else {
      output += pipes
    }

		output += GetEventRgbAnsiSeq(entry.event, instance, false) + "└🡲 \033[1m" + endFmt + " \033[22;2;9m" + entry.event.Props.Summary + "\033[0m"
	}
	return output
}

func displayPipes(instance *Instance, entry *eventEntry) string {
	pipes := []string{}
	parent := entry.parent
	for parent != nil && parent.event != nil {
		pipes = append(pipes, GetEventRgbAnsiSeq(parent.event, instance, false)+"│ \033[0m")
		parent = parent.parent
	}
	slices.Reverse(pipes)
	return strings.Join(pipes, "")
}

func DisplayTimeline(instance *Instance, events []Event, showDates bool, lastShownDate time.Time, location *time.Location) string {
	// Sort the events first:

	slices.SortFunc(events, func(e1 Event, e2 Event) int {
		if s := e1.Props.Start.Compare(e2.Props.Start); s != 0 {
			return s
		} else {
			return e2.Props.End.Compare(e1.Props.End)
		}
	})

	// Put the events into a tree:

	rootEntry := eventEntry{
		children: []*eventEntry{},
	}

	for _, event := range events {
		cursor := &rootEntry
		for len(cursor.children) > 0 {
			lastChild := cursor.children[len(cursor.children)-1]
			if !IsPeriodConfinedToPeriod(event.Props.GetTimeRange(), lastChild.event.Props.GetTimeRange()) {
				break
			}
			cursor = lastChild
		}

		cursor.children = append(cursor.children, &eventEntry{
			event:    &event,
			parent:   cursor,
			children: []*eventEntry{},
		})
	}

	// Display the tree:

	return displayEntry(instance, &rootEntry, showDates, &lastShownDate, location)
}

func DisplayUnsatisfiedRecurrences(instance *Instance, unsatisfiedRecurrences []*Event) string {
	var output string

	for _, event := range unsatisfiedRecurrences {
		output += fmt.Sprintf("\n%s\033[1;30m[..]\033[0;2m there are more occurences of '%s'! \033[2mspecify a range to show more.\033[0m", GetEventRgbAnsiSeq(event, instance, true), event.Props.Summary)
	}

	return output
}

func DisplayCalendarLegend(instance *Instance, events []Event) string {
	var output string
	mentionedCals := []string{}

	for _, event := range events {
		cal := event.Path.Calendar()
		if !slices.Contains(mentionedCals, cal) {
			output += fmt.Sprintf(GetEventRgbAnsiSeq(&event, instance, false)+"▆ %s\033[0m\n", cal)
			mentionedCals = append(mentionedCals, cal)
		}
	}

	return output
}
