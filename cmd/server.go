package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/truecrunchyfrog/ian"
	"github.com/truecrunchyfrog/ian/server"
)

var serverAddressNative string
var serverAddressCalDav string
var serverDebugMode bool

func init() {
  serverCmd.Flags().StringVarP(&serverAddressNative, "address-ian", "i", ":5545", "The server address. Prefix with a semicolon like: ':5545', to just provide a port.")
  serverCmd.Flags().StringVarP(&serverAddressCalDav, "address-caldav", "a", ":80", "The server address. Prefix with a semicolon like: ':5545', to just provide a port.")
  serverCmd.Flags().BoolVarP(&serverDebugMode, "debug", "d", false, "Run server in debug mode.")

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "An ian server.",
	Run:   serverCmdRun,
}

func serverCmdRun(cmd *cobra.Command, args []string) {
	instance, err := ian.CreateInstance(GetRoot())
	if err != nil {
		log.Fatal(err)
	}

  server.Run(serverAddressNative, serverAddressCalDav, serverDebugMode, instance)
}
