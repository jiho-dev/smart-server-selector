package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/sisyphsu/smart-server-selector/config"
	"github.com/sisyphsu/smart-server-selector/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	KeyConfigFile  = "config"
	KeyConfigDir   = "config-dir"
	KeyHostFile    = "host-file"
	KeySshKeyFile1 = "ssh-key-file1"
	KeySshKeyFile2 = "ssh-key-file2"
	KeySshKeyFile3 = "ssh-key-file3"
	KeySshKeyFile4 = "ssh-key-file4"
	KeyVerbose     = "verbose"
)

var rootCmd = &cobra.Command{
	Long: `sss - smart-server-selector`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		//os.Exit(1)
		//Main(cmd, args)
	},
}

var menu = &cobra.Command{
	Use:   "menu",
	Short: "Run Flowlogger Service ",
	Run:   Main,
}

////////////////////////////////

func main() {
	rootCmd.Use = os.Args[0]
	rootCmd.Short = os.Args[0]

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func init() {
	cobra.OnInitialize(func() {

		// load from config dir
		if err := readConfig(); err != nil {
			fmt.Printf("unable to read config: %v \n", err)
			os.Exit(1)
		}

		// configure logger
		//level := logrus.InfoLevel

		//log.NewLogger(logConfig)
		log.SetFormatter(&log.JSONFormatter{})

		// show all configs
		/*
			keys := viper.AllKeys()
			sort.Strings(keys)
			for _, key := range keys {
				log.Infof("config %v = %v", key, viper.Get(key))
			}
		*/
	})

	rootCmd.PersistentFlags().String(KeyConfigFile, "", "config file")
	rootCmd.PersistentFlags().String(KeyConfigDir, "", "configure directory")
	rootCmd.PersistentFlags().String(KeyHostFile, "", "Host List File")
	rootCmd.PersistentFlags().String(KeySshKeyFile1, "", "SSH Key File1")
	rootCmd.PersistentFlags().String(KeySshKeyFile2, "", "SSH Key File2")
	rootCmd.PersistentFlags().String(KeySshKeyFile3, "", "SSH Key File3")
	rootCmd.PersistentFlags().String(KeySshKeyFile4, "", "SSH Key File4")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Errorf("failed to bind flags: %v", err)
	}

	rootCmd.AddCommand(menu)
}

func readConfig() error {
	return config.ReadConfig(&config.Option{
		KeyConfigFile: KeyConfigFile,
		KeyConfigDir:  KeyConfigDir,

		DefaultName: "sss",
		DefaultDirs: []string{
			".",
		},
		EnvPrefix: "sss",
	})
}

/////////////////////////

func Main(cmd *cobra.Command, args []string) {
	runewidth.DefaultCondition.EastAsianWidth = false
	app := tview.NewApplication()

	hostFile := viper.GetString(KeyHostFile)
	key1 := viper.GetString(KeySshKeyFile1)
	key2 := viper.GetString(KeySshKeyFile2)
	key3 := viper.GetString(KeySshKeyFile3)
	key4 := viper.GetString(KeySshKeyFile4)

	var skey string
	if len(args) > 0 {
		skey = strings.Join(args, " ")
	}

	var showAbout bool = false
	cfg := selector.SshConfig{
		SearchKey: skey,
		HostFile:  hostFile,
		KeyFile: map[string]string{
			"dev2":  key1,
			"stg2":  key2,
			"ppd2":  key3,
			"spceu": key4,
		},
	}

	selector.Start(cfg, showAbout, app)

	if err := app.Run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
