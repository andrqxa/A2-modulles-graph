package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func ConfigViper() {
	// Set configuration file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	// Read configuration file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file")
		os.Exit(1)
	}
}
