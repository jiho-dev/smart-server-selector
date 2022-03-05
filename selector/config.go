package selector

import (
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var SidebarWidth = 23

var (
	KeyConfigFile = "config"
	KeyConfigDir  = "config-dir"
	KeyHostFile   = "host-file"
	KeyShowAbout  = "show-about"
	KeyShowBadge  = "show-badge"
	KeySshKeyFile = "ssh-key"
	KeyUserName   = "user-name"
	KeySshPort    = "ssh-port"
	KeySshArgs    = "ssh-args"
)

type SssConfig struct {
	HostFile    string
	ShowBadge   bool // Show hostname+ip as Badge in iterms2
	ShowAbout   bool
	UseSshShell bool
	KeyFile     map[string]string // key: env, data: ssh key file
	SshPort     map[string]string // key: env, data: ssh port
	UserName    map[string]string // key: env, data: user name
	SshArgs     string
}

var SssCfg *SssConfig

func ExpandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}

func getConfigItem(key string) map[string]string {
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

func GetConfig() *SssConfig {
	hostFile := viper.GetString(KeyHostFile)
	hostFile = ExpandPath(hostFile)

	sshKeyFile := getConfigItem(KeySshKeyFile)
	for k, v := range sshKeyFile {
		sshKeyFile[k] = ExpandPath(v)
	}

	userName := getConfigItem(KeyUserName)
	sshPort := getConfigItem(KeySshPort)
	showAbout := viper.GetBool(KeyShowAbout)
	showBadge := viper.GetBool(KeyShowBadge)
	sshArgs := viper.GetString(KeySshArgs)

	cfg := &SssConfig{
		HostFile:  hostFile,
		ShowAbout: showAbout,
		ShowBadge: showBadge,
		KeyFile:   sshKeyFile,
		UserName:  userName,
		SshPort:   sshPort,
		SshArgs:   sshArgs,
	}

	return cfg
}
