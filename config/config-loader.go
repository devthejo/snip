package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/devthejo/snip/errors"
	"github.com/devthejo/snip/goenv"
	"github.com/devthejo/snip/tools"
)

type ConfigLoader struct {
	EnvPrefix                string
	Viper                    *viper.Viper
	ViperDecoderConfigOption viper.DecoderConfigOption
	File                     *string
	Config                   *Config
	RootCmd                  *cobra.Command
	configPaths              []string
	configName               string
	configFile               string
	initialized              bool
}

func NewConfigLoader() *ConfigLoader {
	cl := &ConfigLoader{}
	cl.Viper = viper.New()
	cl.Config = &Config{}
	cl.configPaths = []string{".", "/etc"}
	cl.configName = "snip"
	return cl
}

func (cl *ConfigLoader) ConfigShouldLoad() bool {
	if len(os.Args) < 1 {
		return false
	}
	switch os.Args[1] {
	case "completion":
		return false
	default:
		return true
	}
}

func (cl *ConfigLoader) OnInitialize() {
	if cl.initialized {
		return
	}
	cl.initialized = true

	if !cl.ConfigShouldLoad() {
		return
	}

	cl.InitConfig()

	cl.Load()
}

func (cl *ConfigLoader) OnPreRun(cmd *cobra.Command) {
	if !cl.ConfigShouldLoad() {
		return
	}

	v := cl.Viper

	flags := cmd.Flags()

	var cmdNames []string
	cmd.VisitParents(func(cmd *cobra.Command) {
		cmdNames = append(cmdNames, tools.KeyEnv(cmd.Name()))
	})
	if len(cmdNames) > 0 {
		cmdNames = cmdNames[0 : len(cmdNames)-1]
	}
	cmdNames = append(cmdNames, tools.KeyEnv(cmd.Name()))
	cmdName := strings.Join(cmdNames, "_")

	// v.BindPFlags(flags)
	flags.VisitAll(func(flag *pflag.Flag) {
		envKey := tools.KeyEnv(flag.Name)
		v.BindPFlag(envKey, flags.Lookup(flag.Name))

		l := len(cmdName) + 1
		if len(envKey) < l || envKey[0:l] != cmdName+"_" {
			envKey2 := cmdName + "_" + envKey
			v.BindPFlag(envKey2, flags.Lookup(flag.Name))
		}
	})

	cl.Load()
}

func (cl *ConfigLoader) Load() {
	cl.LoadViper()
	ConfigureLogrusLogType(cl.Config.LogType, cl.Config.LogForceColors)
	ConfigureLogrusLogLevel(cl.Config.LogLevel)
}

func (cl *ConfigLoader) InitConfig() {
	cl.ConfigLogFromEnv()
	cl.ConfigureCWD()
	cl.LoadDotEnv()
	cl.ConfigLogFromEnv()
	cl.LoadJsonnet()
	cl.InitViper()
	cl.LoadViperConfigFile()
	cl.ConfigureDeploymentName()
}

func (cl *ConfigLoader) GetEnvPrefix() string {
	return cl.EnvPrefix
}
func (cl *ConfigLoader) SetEnvPrefix(envPrefix string) {
	cl.EnvPrefix = envPrefix
}
func (cl *ConfigLoader) GetViper() *viper.Viper {
	return cl.Viper
}
func (cl *ConfigLoader) SetViper(v *viper.Viper) {
	cl.Viper = v
}
func (cl *ConfigLoader) GetFile() *string {
	return cl.File
}
func (cl *ConfigLoader) SetFile(file *string) {
	cl.File = file
}
func (cl *ConfigLoader) GetConfig() *Config {
	return cl.Config
}
func (cl *ConfigLoader) SetConfig(config *Config) {
	cl.Config = config
}

func (cl *ConfigLoader) PrefixEnv(key string) string {
	return cl.EnvPrefix + "_" + key
}

func (cl *ConfigLoader) loadDotEnvFile(envfile string) {
	pwd, _ := os.Getwd()
	if ok, err := tools.FileExists(pwd + "/" + envfile); ok {
		logrus.Debugf("loading %v into env", envfile)
		goenv.Load(envfile)
	} else if err != nil {
		logrus.Fatal(err)
	}
}
func (cl *ConfigLoader) LoadDotEnv() {
	cl.loadDotEnvFile(".env.default")
	cl.loadDotEnvFile(".env")

	envKey := cl.PrefixEnv("ENV")
	snipEnv := os.Getenv(envKey)
	logrus.Debugf("snip_env: %v", snipEnv)
	envs := strings.Split(snipEnv, ",")
	for _, env := range envs {
		cl.loadDotEnvFile(".env." + env)
	}

}

func (cl *ConfigLoader) LoadJsonnet() {
	File := *cl.File
	var configDirs []string
	var configName string
	if File != "" {
		dir := filepath.Dir(File)
		ext := filepath.Ext(File)
		base := filepath.Base(File)
		configDirs = []string{dir}
		configName = base[:len(base)-len(ext)]
		if ext == ".jsonnet" {
			*cl.File = dir + "/" + configName + ".json"
		}
	} else {
		configDirs = cl.configPaths
		configName = cl.configName
	}
	if _, err := ConfigJsonnetRender(configDirs, configName); err != nil {
		logrus.Fatal(err)
	}
}

func (cl *ConfigLoader) ConfigLogFromEnv() {
	flags := cl.RootCmd.PersistentFlags()

	logType := cl.GetOptionString(flags, "log-type", FlagLogTypeDefault)
	logForceColors := cl.GetOptionBool(flags, "log-force-colors", FlagLogForceColorsDefault)
	ConfigureLogrusLogType(logType, logForceColors)

	logLevel := cl.GetOptionString(flags, "log-level", FlagLogLevelDefault)
	ConfigureLogrusLogLevel(logLevel)
}

func (cl *ConfigLoader) InitViper() {
	v := cl.Viper
	v.AutomaticEnv()
	v.AllowEmptyEnv(false)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.SetEnvPrefix(cl.EnvPrefix)

	File := *cl.File
	if File != "" {
		v.SetConfigFile(File)
	} else {
		for _, configPath := range cl.configPaths {
			v.AddConfigPath(configPath)
		}
		v.SetConfigName(cl.configName)
	}

	cl.ViperDecoderConfigOption = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		DecodeHookParseDuration(),
		DecodeHookJsonStringAutoDecode(*cl.Config),
		mapstructure.StringToSliceHookFunc(","),
	))
}

func (cl *ConfigLoader) LoadViperConfigFile() {
	if err := cl.Viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logrus.Fatalf("Unable to read config: %v", err)
		}
	}
}

func (cl *ConfigLoader) LoadViper() {
	mergeConfig := &Config{}
	if err := cl.Viper.Unmarshal(mergeConfig, cl.ViperDecoderConfigOption); err != nil {
		logrus.Fatalf("Unable to unmarshal config: %v", err)
	}
	err := mergo.Merge(cl.Config, *mergeConfig, mergo.WithOverride)
	errors.Check(err)
}

func (cl *ConfigLoader) GetOptionString(flags *pflag.FlagSet, key string, defaultValue string) string {
	keyEnv := tools.KeyEnv(key)
	var str string
	if flags.Changed(key) {
		str, _ = flags.GetString(key)
	} else {
		str = os.Getenv(cl.PrefixEnv(keyEnv))
	}
	if str == "" {
		str = defaultValue
	}
	return str
}

func (cl *ConfigLoader) GetOptionBool(flags *pflag.FlagSet, key string, defaultValue bool) bool {
	keyEnv := tools.KeyEnv(key)
	var b bool
	if flags.Changed(key) {
		b, _ = flags.GetBool(key)
	} else {
		s := os.Getenv(cl.PrefixEnv(keyEnv))
		if s == "true" || s == "1" {
			b = true
		} else if s == "false" || s == "0" {
			b = false
		} else {
			b = defaultValue
		}
	}
	return b
}
