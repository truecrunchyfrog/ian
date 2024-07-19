package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

func init() {
	eventPropsCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:     "info event...",
	Aliases: []string{"about", "i"},
	Short:   "View event(s) info",
	Run:     infoCmdRun,
}

func infoCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

	events, _, err := instance.ReadEvents(ian.TimeRange{})
	if err != nil {
		log.Fatal(err)
	}

	infoEvents := []*ian.Event{}

	if len(args) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			args = append(args, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	for _, arg := range args {
		event, err := ian.GetEvent(&events, arg)
		if err != nil {
			log.Fatal(err)
		}
		infoEvents = append(infoEvents, event)
	}

	for _, event := range infoEvents {
		pairs := []struct {
			key   string
			value any
		}{
			{"path", event.Path},
			{"constant", event.Constant},
			{"parent", event.Parent},
			{"", ""},
			{"summary", event.Props.Summary},
			{"start", event.Props.Start},
			{"end", event.Props.End},
			{"duration", ian.DurationToString(event.Props.End.Sub(event.Props.Start))},
			{"recurrence", event.Props.Recurrence},
			{"", ""},
			{"description", event.Props.Description},
			{"location", event.Props.Location},
			{"url", event.Props.Url},
			{"", ""},
			{"created", event.Props.Created},
			{"modified", event.Props.Modified},
			{"", ""},
			{"uid", event.Props.Uid},
		}

		for _, keyValue := range pairs {
			fmt.Println(DisplayKeyValue(keyValue.key, keyValue.value))
		}
	}
}

func DisplayKeyValue(key string, value any) string {
	return fmt.Sprintf("\033[2m%-10s\033[0m %v", key, value)
}
