package config

import (
	"github.com/spf13/viper"
)

func ConfigViper() error {
	// Set configuration file
	viper.SetConfigName("config")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")

	// Read configuration file
	return viper.ReadInConfig()
}
