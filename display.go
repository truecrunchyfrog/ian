package ian

import (
	"fmt"
	"image/color"
	"path"
	"slices"
	"strings"
	"time"
)

func DisplayCalendar(location *time.Location, year int, month time.Month, today time.Time, sunday, showWeeks bool, format func(monthDay int, isToday bool) string) (output string) {
	var weekdayOffset time.Weekday = 1
	if sunday {
		weekdayOffset = 0
	}

	header := fmt.Sprintf("%s %d", month.String(), year)

	if showWeeks {
		output += "    "
	}

	width := 4 * 7 // 4 chars per day, 7 days in a week
	output += fmt.Sprintf("%[1]*s\n", -width, fmt.Sprintf("%[1]*s", (width+len(header))/2, header))

	daysInMonth := 32 - time.Date(year, month, 32, 0, 0, 0, 0, location).Day()
	firstWeekdayInMonth := time.Date(year, month, 1, 0, 0, 0, 0, location).Weekday()

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
			_, week := time.Date(year, month, 1, 0, 0, 0, 0, location).ISOWeek()
			output += displayWeek(week)
		}
		output += strings.Repeat(" ", 4*emptyDays)
	}

	for monthDay := 1; monthDay <= daysInMonth; monthDay++ {
		weekday := time.Weekday((int(firstWeekdayInMonth) + monthDay - 1) % 7)

		if showWeeks && weekday == weekdayOffset {
			_, week := time.Date(year, month, monthDay, 0, 0, 0, 0, location).ISOWeek()
			output += displayWeek(week)
		}

		isToday := year == today.Year() && month == today.Month() && monthDay == today.Day()

		format := format(monthDay, isToday)
		output += fmt.Sprintf(" "+format+"%3s\033[0m", fmt.Sprint(monthDay))

		if weekday == (weekdayOffset+6)%7 {
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
func possibleEntryDate(now time.Time, lastShownDate *time.Time) string {
	output := ""

	if now.YearDay() != lastShownDate.YearDay() || now.Year() != lastShownDate.Year() {
		output += now.Format("_2 Jan")
		*lastShownDate = now
	}

	return fmt.Sprintf("\033[2m%-8s\033[22m", output)
}

// TODO add parallel support
func displayEntry(instance *Instance, entry *eventEntry, lastShownDate *time.Time, location *time.Location) string {
	var output string

	var period string
	var startFmt, endFmt string
	if entry.event != nil {
		// Itself

		start := entry.event.Props.Start.In(location)
		end := entry.event.Props.End.In(location)
		if !entry.event.Props.AllDay {
			startFmt = start.Format("15")
			if start.Minute() == 0 {
				//startFmt += "\033[2m"
			}
			if start.Minute() != 0 {
				startFmt += start.Format(":04") + "\033[22m"
			}

			endFmt = end.Format("15")
			if end.Minute() == 0 {
				//endFmt += "\033[2m"
			}
			if end.Minute() != 0 {
				endFmt += end.Format(":04") + "\033[22m"
			}

			if len(entry.children) != 0 {
				period = startFmt
			} else {
				//onlyToday := start.Day() == end.Day()
				period = startFmt + " 🡲  " + endFmt
			}
		} else {
			period = "*"
		}

    // TODO wrap text
		output += possibleEntryDate(start, lastShownDate) + displayPipes(instance, entry) + GetEventRgbAnsiSeq(entry.event, instance, false) + "\033[1m" + period + "\033[22m " + entry.event.Props.Summary + "\033[0m"
	}
	for _, child := range entry.children {
		// Children
		output += "\n" + displayEntry(instance, child, lastShownDate, location)
	}
	if entry.event != nil && len(entry.children) != 0 {
		// Tail
		output += "\n" + possibleEntryDate(entry.event.Props.End.In(location), lastShownDate) + displayPipes(instance, entry) + GetEventRgbAnsiSeq(entry.event, instance, false) + "└🡲 \033[1m" + endFmt + "\033[22m " + entry.event.Props.Summary + "\033[0m"
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

func DisplayTimeline(instance *Instance, periodStart, periodEnd time.Time, events []Event, location *time.Location) string {
	// Sort the events first:

	slices.SortFunc(events, func(e1 Event, e2 Event) int {
		if s := e1.Props.Start.Compare(e2.Props.Start); s != 0 {
			return s
		} else {
			return e2.Props.End.Compare(e1.Props.End) // NOTE these (e1 and e2) were switched in attempt to make the later end date come first, for visual clustering!
		}
	})

	// Put the events into a tree:

	// Example:
	/*

	   1 Jan 09:00 Project
	   2 Jan │ 15:00→18:40 Discuss project with team
	   3 Jan │ 09:00 Workshop
	         │ │ 10:30→13:00 Work on project with team
	         │ └16:00 end <Workshop>
	   7 Jan │ 17:00→19:00 Event in the middle
	  14 Jan └15:00 end <Project>

	*/

	rootEntry := eventEntry{
		children: []*eventEntry{},
	}

	for _, event := range events {
		start := event.Props.Start.In(location)
		end := event.Props.End.In(location)

		cursor := &rootEntry
		for len(cursor.children) > 0 {
			lastChild := cursor.children[len(cursor.children)-1]
			if !IsPeriodConfinedToPeriod(start, end, lastChild.event.Props.Start, lastChild.event.Props.End) {
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

	//cursor := periodStart
	//ongoingEvents := []Event{}

	/*for _, event := range events {
	  start := event.Props.Start.In(location)
	  end := event.Props.End.In(location)
	  for _, ongoingEvent := range ongoingEvents {
	    if ongoingEvent.Props.End.Before()
	  }
	  if start.Year() == end.Year() && start.YearDay() == end.YearDay() {
	    // Event is during a single day.
	    if event.Props.AllDay {
	      // It's an all-day event. Display no time.
	      output += ""
	    } else {
	      // It's during a part of a single day. Display the time.
	    }
	  }
	}*/
}
