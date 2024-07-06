package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
)

var clean bool
var updateAll bool
var updateSources []string

func init() {
	sourcesCmd.Flags().BoolVarP(&clean, "clean", "c", false, "Delete and update all sources.")
	sourcesCmd.Flags().BoolVarP(&updateAll, "update-all", "U", false, "Update all lists.")
	sourcesCmd.Flags().StringSliceVarP(&updateSources, "update", "u", []string{}, "Update a list of comma-separated source `names`. E.g.: 'ian sources --update school,home'.")
	sourcesCmd.MarkFlagsMutuallyExclusive("clean", "update-all", "update")

	rootCmd.AddCommand(sourcesCmd)
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Manage and view sources",
	Args:  cobra.NoArgs,
	Run:   sourcesCmdRun,
}

func sourcesCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot(), false)
	if err != nil {
		log.Fatal(err)
	}

  if clean {
    log.Println("cleaning...")
    if err := instance.CleanSources(); err != nil {
      log.Fatal(err)
    }
  }

  for _, name := range updateSources {
    if _, ok := instance.Config.Sources[name]; !ok {
      log.Fatalf("no such source: '%s'\n", name)
    }
  }

	for name, source := range instance.Config.Sources {
		fmt.Printf("%s:\n\t%s\n", name, source.Source)

		if updateAll || slices.Contains(updateSources, name) {
			fmt.Printf("(updating '%s'...)\n", name)
      if err := source.ImportAndUse(instance, name); err != nil {
        log.Fatal(err)
      }
		}

    fmt.Println()
	}
}
