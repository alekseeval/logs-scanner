package configuration

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	System struct {
		Http struct {
			Port    int
			Timeout int
		}
		Postgres struct {
			Ip       string
			Port     int
			DbName   string
			User     string
			Password string
			Timeout  int
		}
		Kubernetes struct {
			Timeout int
		}
	}
	Logger struct {
		LogLevel string
	}
	ScanDelay      int
	JobGrepPattern string
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
