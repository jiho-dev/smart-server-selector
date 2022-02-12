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

var (
	KeyConfigFile = "config"
	KeyConfigDir  = "config-dir"
	KeyHostFile   = "host-file"
	KeyShowAbout  = "show-about"
	KeySshKeyFile = "ssh-key"
	KeyUserName   = "user-name"
	KeySshPort    = "ssh-port"
	KeyVerbose    = "verbose"
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

	rootCmd.PersistentFlags().String(KeyConfigFile, cfgFile, "config file")
	rootCmd.PersistentFlags().String(KeyConfigDir, "", "configure directory")
	rootCmd.PersistentFlags().String(KeyHostFile, hostFile, "Host List File")
	rootCmd.PersistentFlags().Bool(KeyShowAbout, false, "Show About Menu")
	rootCmd.PersistentFlags().String(KeySshKeyFile, "", "SSH Key File1")

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

func GetConfig(key string) map[string]string {
	values := map[string]string{}

	vals := viper.Get(key)
	val, ok := vals.([]interface{})
	if ok {
		for _, val2 := range val {
			val3, ok3 := val2.(map[interface{}]interface{})
			if ok3 {
				for key4, val4 := range val3 {
					k := key4.(string)
					v := val4.(string)
					values[k] = v
				}
			}
		}
	}

	return values
}

func Main(cmd *cobra.Command, args []string) {
	runewidth.DefaultCondition.EastAsianWidth = false
	app := tview.NewApplication()

	hostFile := viper.GetString(KeyHostFile)
	sshKeyFile := GetConfig(KeySshKeyFile)
	userName := GetConfig(KeyUserName)
	sshPort := GetConfig(KeySshPort)

	var skey string
	if len(args) > 0 {
		skey = strings.Join(args, " ")
	}

	showAbout := viper.GetBool(KeyShowAbout)
	cfg := selector.SshConfig{
		SearchKey: skey,
		HostFile:  hostFile,
		KeyFile:   sshKeyFile,
		UserName:  userName,
		SshPort:   sshPort,
	}

	selector.Start(&cfg, showAbout, app)

	if err := app.Run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
