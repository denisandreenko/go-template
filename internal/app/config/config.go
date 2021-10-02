package config

import (
	"github.com/spf13/viper"
)

type (
	AppConfig struct {
		HTTP struct {
			Host  string
			Port  int
			Check struct {
				Host string
			}
		}
		GRPC struct {
			Host  string
			Port  int
			Check struct {
				Host string
			}
		}
		Consul struct {
			Disabled bool
			Scheme   string
			Host     string
			Port     int
			DC       string
			Check    struct {
				Interval string
			}
		}
		ApmServer struct {
			Name string
			Tag  string
			DC   string
		}
	}
)

func NewAppConfig(configFile string) (*AppConfig, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadConfig(configFile string) (*AppConfig, error) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var appConfig AppConfig
	err = viper.Unmarshal(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, err
}
