package main

import (
	"log"

	"github.com/truecrunchyfrog/ian/cmd"
)

func main() {
	log.SetFlags(0)
  log.SetPrefix("ian: ")

	cmd.Execute()
}
