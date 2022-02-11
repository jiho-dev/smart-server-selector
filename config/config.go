package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Option struct {
	// if provided, try to read config file at first order
	KeyConfigFile string
	// if provided, try to read config files from directory
	KeyConfigDir string

	// if failed to discover config file, try to search following directories
	DefaultName string
	DefaultDirs []string

	// if provided, try to read config from environment
	EnvPrefix string

	// default values
	Defaults map[string]interface{}
}

// protect viper internal map
var lock sync.RWMutex

func ReadConfig(opt *Option) error {
	var (
		cfgFile string
		cfgDir  string
	)

	lock.Lock()
	defer lock.Unlock()

	// set defaults
	for k, v := range opt.Defaults {
		viper.SetDefault(k, v)
	}

	// try to read config from file
	if opt.KeyConfigFile != "" {
		cfgFile = viper.GetString(opt.KeyConfigFile)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(opt.DefaultName)
		for _, p := range opt.DefaultDirs {
			viper.AddConfigPath(p)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return errors.Wrap(err, "viper.ReadInConfig")
		}
	}

	// try to read config from directory
	if opt.KeyConfigDir != "" {
		cfgDir = viper.GetString(opt.KeyConfigDir)
	}

	if cfgDir != "" {
		if err := readConfigDir(cfgDir); err != nil {
			return errors.Wrap(err, "readConfigDir")
		}
	}

	//
	viper.AutomaticEnv()
	if opt.EnvPrefix != "" {
		viper.SetEnvPrefix(opt.EnvPrefix)
	}
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	return nil
}

func readConfigFile(filename string) error {
	finfo, err := os.Stat(filename)
	if err != nil {
		return errors.Wrapf(err, "os.Stat(%v)", filename)
	}

	if finfo.IsDir() {
		return errors.Wrapf(err, "invalid input: %v is directory", filename)
	}

	var b []byte

	if finfo.Mode().Perm()&0111 > 0 {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cmd := exec.Command("/bin/sh", "-c", filename)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "cmd.Run(%v), stdout: %v stderr: %v",
				filename, stdout.String(), stderr.String())
		}

		viper.SetConfigType("json")
		b = stdout.Bytes()
	} else {
		switch path.Ext(filename) {
		case ".yaml", ".yml":
			viper.SetConfigType("yaml")
		case ".json", ".js":
			viper.SetConfigType("json")
		default:
			return errors.Errorf("unsupported file extention: %v", path.Ext(filename))
		}

		b, err = ioutil.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "ioutil.ReadFile(%v)", filename)
		}

	}

	if err := viper.MergeConfig(bytes.NewBuffer(b)); err != nil {
		return errors.Wrap(err, "viper.MergeConfig")
	}

	return nil
}

func readConfigDir(dirname string) error {
	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		return errors.Wrapf(err, "unable to read directory(%v)", dirname)
	}

	for i := range entries {
		if entries[i].IsDir() {
			continue
		}

		filepath := path.Join(dirname, entries[i].Name())
		if err := readConfigFile(filepath); err != nil {
			return errors.Wrapf(err, "readConfigFile(%v)", filepath)
		}
	}

	return nil
}

func BindPFlag(key string, flag *pflag.Flag) error {
	lock.Lock()
	defer lock.Unlock()
	return viper.BindPFlag(key, flag)
}

func BindPFlags(flags *pflag.FlagSet) error {
	lock.Lock()
	defer lock.Unlock()
	return viper.BindPFlags(flags)
}

func BindKey(key, d string) {
	lock.Lock()
	defer lock.Unlock()
	viper.SetDefault(key, d)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}

func Get(key string) string {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetString(key)
}

func GetBool(key string) bool {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetBool(key)
}

func GetInt(key string) int {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetInt(key)
}

func GetInt32(key string) int32 {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetInt32(key)
}

func GetFloat64(key string) float64 {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetFloat64(key)
}

func GetInterface(key string) interface{} {
	lock.RLock()
	defer lock.RUnlock()
	return viper.Get(key)
}

func GetDuration(key string) time.Duration {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetDuration(key)
}

func GetStringMapString(key string) map[string]string {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetStringMapString(key)
}

func GetStringMap(key string) map[string]interface{} {
	lock.RLock()
	defer lock.RUnlock()
	return viper.GetStringMap(key)
}
