package main

import (
	"fmt"
	"os"

	"github.com/sisyphsu/smart-server-selector/selector"
	"github.com/spf13/cobra"
)

var conHostCmd = &cobra.Command{
	Use:     "connect",
	Aliases: []string{"c", "con", "conn"},
	Short:   "Connect Host directly ",
	Run:     connectHost,
}

////////////////////////////////

func init() {
	rootCmd.AddCommand(conHostCmd)
}

/////////////////////////

func connectHost(cmd *cobra.Command, args []string) {
	var host string
	if len(args) > 0 {
		host = args[0]
	}

	cfg := selector.GetConfig()

	err := selector.StartSSHExt(cfg, host)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

}
