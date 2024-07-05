package main

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var duration string

func init() {
	addCommand.Flags().StringVarP(&duration, "duration", "d", "", "Duration of the event from start to end. Use instead of providing the 'end' argument. Example: --duration 1h30m to set 'end' to 'start' with 1 hour and 30 minutes added.")

	rootCmd.AddCommand(addCommand)
}

var addCommand = &cobra.Command{
	Use:   "add <description> <start> [end]",
	Short: "Create a new event",
	Long:  "The arguments 'start' and 'end' support many different formats. The recommended format for most events is dd/mm, or dd/mm hh:mm with time. A time zone can be appended with +-hhmm or 'UTC'-like format.",
	Args:  cobra.RangeArgs(2, 3),
	Run:   addCommandRun,
}

func addCommandRun(cmd *cobra.Command, args []string) {
	if len(args) >= 3 && duration != "" {
		log.Fatal("'end' and '--duration' ('-d') cannot be combined")
	}

	tz := GetTimeZone()

	startDate, err := ian.ParseDateTime(args[1], &tz)
	if err != nil {
		log.Fatal(err)
	}

	var endDate time.Time
	if len(args) >= 3 {
		var err error
		endDate, err = ian.ParseDateTime(args[2], &tz)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if duration != "" {
			d, err := time.ParseDuration(duration)
			if err != nil {
				log.Fatal(err)
			}
			endDate = startDate.Add(d)
		} else {
			endDate = startDate.Add(24 * time.Hour)
		}
	}

	props := ian.EventProperties{
		Summary:  args[0],
		Start: startDate,
		End:   endDate,
	}

  if err := props.Verify(); err != nil {
    log.Fatal("invalid event: ", err)
  }

  instance, err := ian.CreateInstance(GetRoot())
  if err != nil {
    log.Fatal(err)
  }

  instance.CreateEvent(props, "")

	if err := instance; err != nil {
		log.Fatal(err)
	}

	eventDuration := endDate.Sub(startDate)

	log.Printf("%s\n\n%s (%s)\n%s\n",
		props.Summary,
		props.Start.Format(ian.DefaultTimeLayout),
		eventDuration.String(),
		props.End.Format(ian.DefaultTimeLayout),
	)
}
