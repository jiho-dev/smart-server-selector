package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/sisyphsu/smart-server-selector/config"
	"github.com/sisyphsu/smart-server-selector/selector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Use:     "menu",
	Aliases: []string{"m", "me", "men"},
	Short:   "Show host selector",
	Run:     Main,
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

		/*
			// show all configs
			keys := viper.AllKeys()
			sort.Strings(keys)
			for _, key := range keys {
				log.Infof("config %v = %v", key, viper.Get(key))
			}
		*/
	})

	homeDir, _ := os.UserHomeDir()
	cfgFile := path.Join(homeDir, ".ssh", "sss.yaml")
	hostFile := path.Join(homeDir, ".ssh", "sss-host.cfg")

	rootCmd.PersistentFlags().String(selector.KeyConfigFile, cfgFile, "config file")
	rootCmd.PersistentFlags().String(selector.KeyConfigDir, "", "configure directory")
	rootCmd.PersistentFlags().String(selector.KeyHostFile, hostFile, "Host List File")
	rootCmd.PersistentFlags().Bool(selector.KeyShowAbout, false, "Show About Menu")
	rootCmd.PersistentFlags().Bool(selector.KeyShowBadge, true, "Show Badge")
	rootCmd.PersistentFlags().String(selector.KeySshKeyFile, "", "SSH Key File")
	rootCmd.PersistentFlags().String(selector.KeyUserName, "", "Default User Name")
	rootCmd.PersistentFlags().String(selector.KeySshPort, "", "Default SSH Port")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Errorf("failed to bind flags: %v", err)
	}

	rootCmd.AddCommand(menu)
}

func readConfig() error {
	return config.ReadConfig(&config.Option{
		KeyConfigFile: selector.KeyConfigFile,
		KeyConfigDir:  selector.KeyConfigDir,

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

	var skey string
	if len(args) > 0 {
		skey = strings.Join(args, " ")
	}

	cfg := selector.GetConfig()
	selector.Start(cfg, skey, app)

	if err := app.Run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
