package cmd

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	findCmd.Flags().StringSliceP("calendars", "c", nil, "Events must be located in a calendar listed here.")
	findCmd.Flags().BoolP("ignore-constant", "i", false, "Ignore constant events (cache and recurrences).")
	findCmd.Flags().Bool("case-sensitive", false, "Query matching is sensitive to casing.")
	findCmd.Flags().StringP("path", "p", "", "Query the events' paths.")
	findCmd.Flags().StringP("summary", "s", "", "Query the events' summary.")
	findCmd.Flags().StringP("description", "d", "", "Query the events' description.")

	findCmd.Flags().StringSlice("at", nil, "Events must occur during this/these datetime(s).")
	findCmd.Flags().String("before", "", "Events must end before this datetime.")
	findCmd.Flags().String("after", "", "Events must start after this datetime.")
	findCmd.Flags().Bool("exclusive", false, "When combined with 'before' and/or 'after', the entire event time ranges must occur outside of these limits (e.g., if 'before' is set to 01-04-1991, then an event cannot start before and end after 1 April; the entire time range must be confined before that datetime).")
  findCmd.Flags().BoolP("one", "1", false, "Exit if the query does not match exactly one event.")

	findCmd.MarkFlagsMutuallyExclusive("at", "before")
	findCmd.MarkFlagsMutuallyExclusive("at", "after")

	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:     "find [-p path] [-s summary] [--before date] [--after date]",
	Aliases: []string{"f", "q", "l", "list"},
	Short:   "Query events",
	Args:    cobra.NoArgs,
	Run:     findCmdRun,
}

func findCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	calendars, _ := cmd.Flags().GetStringSlice("calendars")
	ignoreConstant, _ := cmd.Flags().GetBool("ignore-constant")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	path, _ := cmd.Flags().GetString("path")
	summary, _ := cmd.Flags().GetString("summary")
	description, _ := cmd.Flags().GetString("description")

	occurAt := []time.Time{}
	occurBefore := time.Time{}
	occurAfter := time.Time{}
	exclusiveBeforeAfter, _ := cmd.Flags().GetBool("exclusive")

	if at, _ := cmd.Flags().GetStringSlice("at"); len(at) != 0 {
		for _, entryAt := range at {
			t, err := ian.ParseDateTime(entryAt, ian.GetTimeZone())
			if err != nil {
				log.Fatal(err)
			}
			occurAt = append(occurAt, t)
		}
	}

	if before, _ := cmd.Flags().GetString("before"); before != "" {
		occurBefore, err = ian.ParseDateTime(before, ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	}

	if after, _ := cmd.Flags().GetString("after"); after != "" {
		occurAfter, err = ian.ParseDateTime(after, ian.GetTimeZone())
		if err != nil {
			log.Fatal(err)
		}
	}

	events = ian.FilterEvents(&events, func(e *ian.Event) bool {
		if ignoreConstant && e.Constant {
			return false
		}

		if len(calendars) != 0 && !slices.Contains(calendars, e.Path.Calendar()) {
			return false
		}

		if len(occurAt) != 0 {
			insideOne := false
			for _, at := range occurAt {
				if ian.IsTimeWithinPeriod(at, e.Props.GetTimeRange()) {
					insideOne = true
				}
			}
			if !insideOne {
				return false
			}
		}

		if !occurBefore.IsZero() && (e.Props.Start.After(occurBefore) || (exclusiveBeforeAfter && !e.Props.End.Before(occurBefore))) {
			return false
		}

		if !occurAfter.IsZero() && (e.Props.End.Before(occurAfter) || (exclusiveBeforeAfter && !e.Props.Start.After(occurAfter))) {
			return false
		}

		if path != "" {
			if caseSensitive {
				if !strings.Contains(e.Path.String(), path) {
					return false
				}
			} else if !strings.Contains(strings.ToLower(e.Path.String()), strings.ToLower(path)) {
				return false
			}
		}

		if summary != "" {
			if caseSensitive {
				if !strings.Contains(e.Props.Summary, summary) {
					return false
				}
			} else if !strings.Contains(strings.ToLower(e.Props.Summary), strings.ToLower(summary)) {
				return false
			}
		}

		if description != "" {
			if caseSensitive {
				if !strings.Contains(e.Props.Description, description) {
					return false
				}
			} else if !strings.Contains(strings.ToLower(e.Props.Description), strings.ToLower(description)) {
				return false
			}
		}

		return true
	})

  if one, _ := cmd.Flags().GetBool("one"); one && len(events) != 1 {
    log.Fatalf("expected one result, got %d\n", len(events))
  }

	for _, event := range events {
		fmt.Println(event.Path)
	}
}
