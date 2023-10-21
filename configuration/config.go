package configuration

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	System struct {
		Http struct {
			Port    int `mapstructure:"port"`
			Timeout int `mapstructure:"timeout"`
		} `mapstructure:"http"`
		Postgres struct {
			Ip       string `mapstructure:"ip"`
			Port     int    `mapstructure:"port"`
			DbName   string `mapstructure:"db_name"`
			User     string `mapstructure:"user"`
			Password string `mapstructure:"password"`
			Timeout  int    `mapstructure:"timeout"`
		}
		Kubernetes struct {
			Timeout int `mapstructure:"timeout"`
		}
	} `mapstructure:"system"`
	Logger struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`
	ScanDelay       int    `mapstructure:"scan_delay"`
	JobsGrepPattern string `mapstructure:"jobs_grep_pattern"`
}

func ReadConfig(path string) (config *Config, err error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config)
	return config, err
}
