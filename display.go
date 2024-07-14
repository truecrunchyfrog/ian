package ian

import (
	"fmt"
	"image/color"
	"path"
	"slices"
	"strings"
	"time"
)

func DisplayCalendar(
	location *time.Location,
	year int,
	month time.Month,
	today time.Time,
	sunday,
	showWeeks,
	weekHinting bool,
	borders bool,
	widthPerDay int,
	format func(monthDay int, isToday bool) (string, bool),
) (output string) {
	var weekdayOffset time.Weekday = 1
	if sunday {
		weekdayOffset = 0
	}

	header := fmt.Sprintf("%s %d", month.String(), year)

	if showWeeks {
		output += "    "
	}

	width := (widthPerDay + 1) * 7 // 4 chars per day, 7 days in a week
	output += fmt.Sprintf("%[1]*s\n", -width, fmt.Sprintf("%[1]*s", (width+len(header))/2, header))

	daysInMonth := 32 - time.Date(year, month, 32, 0, 0, 0, 0, location).Day()
	firstWeekdayInMonth := time.Date(year, month, 1, 0, 0, 0, 0, location).Weekday()

	if showWeeks {
		output += "  "
		if borders {
			output += " "
		}
	}

	weekdayFormat := "\033[2m"
	if borders {
		weekdayFormat += "\033[4m"
	}
	output += weekdayFormat
	for i := 0; i < 7; i++ {
		weekday := time.Weekday((int(weekdayOffset) + i) % 7)
		dayString := weekday.String()
		if len(dayString) > widthPerDay {
			dayString = dayString[:widthPerDay]
		}
		if weekHinting && weekday == today.Weekday() {
			dayString = "\033[22m" + dayString + weekdayFormat
		}
		output += fmt.Sprintf(" %"+fmt.Sprint(widthPerDay)+"s", dayString)
	}
	output += "\033[0m\n"

	displayWeek := func(week int) string {
		var format string
		if _, currentWeek := today.ISOWeek(); weekHinting && currentWeek == week {
			format = "\033[22;1;37m"
		}
		var border string
		if borders {
			border = "│"
		}
		return fmt.Sprintf("\033[2m"+format+"%2d"+border+"\033[0m", week)
	}

	emptyDays := int(firstWeekdayInMonth - weekdayOffset)
	if !sunday && firstWeekdayInMonth == 0 {
		emptyDays = 6
	}
	if emptyDays > 0 {
		if showWeeks {
			_, week := time.Date(year, month, 1, 0, 0, 0, 0, location).ISOWeek()
			output += displayWeek(week)
		}
		output += strings.Repeat(" ", (widthPerDay+1)*emptyDays)
	}

	for monthDay := 1; monthDay <= daysInMonth; monthDay++ {
		weekday := time.Weekday((int(firstWeekdayInMonth) + monthDay - 1) % 7)

		if showWeeks && weekday == weekdayOffset {
			_, week := time.Date(year, month, monthDay, 0, 0, 0, 0, location).ISOWeek()
			output += displayWeek(week)
		}

		isToday := year == today.Year() && month == today.Month() && monthDay == today.Day()

		format, entireSlot := format(monthDay, isToday)
		padding := strings.Repeat(" ", widthPerDay-2)
		if widthPerDay > 2 && entireSlot {
			format += padding
		} else {
			format = padding + format
		}
		output += fmt.Sprintf(" "+format+"%s\033[0m", fmt.Sprintf("%2d", monthDay))

		if weekday == (weekdayOffset+6)%7 && monthDay != daysInMonth {
			output += "\n"
		}
	}

	output += "\n"
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
	container := path.Dir(event.Path)
	var rgb color.RGBA
	if conf, err := instance.Config.GetContainerConfig(container); err == nil {
		rgb = conf.GetColor()
	} else {
		rgb = (&ContainerConfig{}).GetColor()
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
	default:
		// Today:
		format = "\033[1;30;47m"
	}
	return fmt.Sprintf("%s%s%s%-6s\033[0m ", year, month, format, date)
}

func displayEntry(instance *Instance, entry *eventEntry, lastShownDate *time.Time, location *time.Location) string {
	var output string

	var period string
	var startFmt, endFmt string
	if entry.event != nil {
		// Itself

		start := entry.event.Props.Start.In(location)
		end := entry.event.Props.End.In(location)
		if !entry.event.Props.AllDay || entry.event.Props.Start.Location() != location {
			startFmt = start.Format("15")
			if start.Minute() != 0 {
				startFmt += start.Format(":04")
			}

			endFmt = end.Format("15")
			if end.Minute() != 0 {
				endFmt += end.Format(":04")
			}

			if len(entry.children) != 0 || start.Day() != end.Day() {
				period = startFmt
			} else {
				period = startFmt + " 🡲  " + endFmt
			}
		} else {
			period = "+"
		}

		pipes := displayPipes(instance, entry)

		entryDate := possibleEntryDate(start, lastShownDate)
		entryDateLines := strings.Split(entryDate, "\n")
		for i := 0; i < len(entryDateLines); i++ {
			entryDateLines[i] += pipes
		}

		output += strings.Join(entryDateLines, "\n") + GetEventRgbAnsiSeq(entry.event, instance, false) + "\033[1m" + period + "\033[22m " + entry.event.Props.Summary + "\033[0m"
	}
	for _, child := range entry.children {
		// Children
		output += "\n" + displayEntry(instance, child, lastShownDate, location)
	}
	if entry.event != nil && (len(entry.children) != 0 || entry.event.Props.Start.In(location).Day() != entry.event.Props.End.In(location).Day()) {
		// Tail

		pipes := displayPipes(instance, entry)
		innerPipes := displayPipes(instance, &eventEntry{parent: entry})

		entryDate := possibleEntryDate(entry.event.Props.End.In(location), lastShownDate)
		entryDateLines := strings.Split(entryDate, "\n")
		for i := 0; i < len(entryDateLines); i++ {
			if i != len(entryDateLines)-1 {
				entryDateLines[i] += innerPipes
			} else {
				entryDateLines[i] += pipes
			}
		}

		output += "\n" + strings.Join(entryDateLines, "\n") + GetEventRgbAnsiSeq(entry.event, instance, false) + "└🡲 \033[1m" + endFmt + " \033[22;2;9m" + entry.event.Props.Summary + "\033[0m"
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

func DisplayTimeline(instance *Instance, events []Event, location *time.Location) string {
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

	lastShownDate := time.Time{}
	return displayEntry(instance, &rootEntry, &lastShownDate, location)
}

func DisplayUnsatisfiedRecurrences(instance *Instance, unsatisfiedRecurrences []*Event) string {
	var output string

	for _, event := range unsatisfiedRecurrences {
		output += fmt.Sprintf("\n%s[..]\033[0;2m there are more occurences of '%s'! \033[2mspecify a range to show more.\033[0m", GetEventRgbAnsiSeq(event, instance, true), event.Props.Summary)
	}

	return output
}

func DisplayCalendarLegend(instance *Instance, events []Event) string {
	var output string
	mentionedCals := []string{}

	for _, event := range events {
		cal := event.GetCalendarName()
		if !slices.Contains(mentionedCals, cal) {
			output += fmt.Sprintf(GetEventRgbAnsiSeq(&event, instance, false)+"▆ %s\033[0m\n", cal)
			mentionedCals = append(mentionedCals, cal)
		}
	}

	return output
}
