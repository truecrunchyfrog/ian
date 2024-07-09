package main

import (
	"log"

	"github.com/truecrunchyfrog/ian/cmd"
)

func main() {
	log.SetFlags(log.Ltime)

	cmd.Execute()
}
