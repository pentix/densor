package main

import (
	"github.com/spf13/viper"
	"os"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string
}

type LocalConfig struct {
	UUID            string
	DisplayName     string
	RemoteInstances []RemoteInstance
}

func readConfig() {
	configDir, err := os.UserConfigDir()
	configPath := configDir + "/densor.yaml"
	if err != nil {
		panic(err)
	}

	// Defaults
	UUID := generateUUID()
	viper.SetDefault("UUID", UUID)
	viper.SetDefault("DisplayName", "Host-"+UUID)
	viper.SetDefault("RemoteInstances", []RemoteInstance{})

	// Try to parse possible existing yaml file or create it
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Println("Error reading the configuration:", err)
	}
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}

	// Read everything into the LocalConfig struct
	config.UUID = viper.GetString("UUID")
	config.DisplayName = viper.GetString("DisplayName")
	err = viper.UnmarshalKey("RemoteInstances", &config.RemoteInstances)
	if err != nil {
		panic(err)
	}

	logger.Println("Number of remote instances: ", len(config.RemoteInstances))
}
