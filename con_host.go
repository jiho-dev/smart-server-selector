package main

import (
	"fmt"
	"os"

	"github.com/sisyphsu/smart-server-selector/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	hostFile := viper.GetString(KeyHostFile)
	key1 := viper.GetString(KeySshKeyFile1)
	key2 := viper.GetString(KeySshKeyFile2)
	key3 := viper.GetString(KeySshKeyFile3)
	key4 := viper.GetString(KeySshKeyFile4)

	var host string
	if len(args) > 0 {
		host = args[0]
	}

	cfg := selector.SshConfig{
		HostFile: hostFile,
		KeyFile: map[string]string{
			"dev2":  key1,
			"stg2":  key2,
			"ppd2":  key3,
			"spceu": key4,
		},
	}

	err := selector.StartSSHExt(cfg, host)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

}
