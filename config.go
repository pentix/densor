package main

import (
	"os"

	"github.com/spf13/viper"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string
}

type LocalConfig struct {
	UUID            string
	DisplayName     string
	DataDir			string
	RemoteInstances []RemoteInstance
	Sensors 		[]Sensor

	config			*viper.Viper
}

func readConfig() {
	// Actual config file
	configDir, err := os.UserConfigDir()
	configPath := configDir + "/densor.yaml"
	if err != nil {
		panic(err)
	}

	// Try to create default data dir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultDataDir := homeDir + "/.densor/"
	os.Mkdir(defaultDataDir, 0644)


	// Default values
	UUID := generateUUID()
	viper.SetDefault("UUID", UUID)
	viper.SetDefault("DisplayName", "Host-"+UUID)
	viper.SetDefault("DataDir", defaultDataDir)
	viper.SetDefault("RemoteInstances", []RemoteInstance{})
	viper.SetDefault("Sensors", []Sensor{})

	// Try to parse possible existing yaml file or create it
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Println("Error reading the configuration:", err)
	}
	if err := viper.WriteConfig(); err != nil {
		panic(err)
	}

	// Read everything into the LocalConfig struct
	config.UUID = viper.GetString("UUID")
	config.DisplayName = viper.GetString("DisplayName")

	if err := viper.UnmarshalKey("RemoteInstances", &config.RemoteInstances); err != nil {
		panic(err)
	}

	if err := viper.UnmarshalKey("Sensors", &config.Sensors); err != nil {
		panic(err)
	}

	// Set viper instance for Sensors access
	config.config = viper.GetViper()

	logger.Println("Number of remote instances: ", len(config.RemoteInstances))
	logger.Println("Number of sensors:          ", len(config.Sensors))
}

func initSensors() {
	// Read measurements
	for _, sensor := range config.Sensors {
		sensor.measurements = viper.New()
		sensor.measurements.SetConfigFile(config.DataDir + sensor.UUID + ".json")
		if err := sensor.measurements.ReadInConfig(); err != nil {
			logger.Printf("Error sensor configuration for sensor %s [%s]\n", sensor.UUID, sensor.DisplayName)
			logger.Printf("%s", err)
			logger.Printf("Skipping initialization of sensor\n")
			continue
		}

		go sensor.enableMeasurements()
	}

	// Wait forever, in an efficient manner
	select {}
}