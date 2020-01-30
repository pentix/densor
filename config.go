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
	DataDir         string
	RemoteInstances map[string]RemoteInstance
	Sensors         map[string]Sensor

	config *viper.Viper
}

func readConfig() {
	// Actual config file
	configDir, err := os.UserConfigDir()
	configPath := configDir + "/densor.json"
	if err != nil {
		panic(err)
	}

	// Try to create default data dir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultDataDir := homeDir + "/.densor/"
	if err := os.Mkdir(defaultDataDir, 0755); !os.IsExist(err) {
		logger.Fatal("Could not create default data directory:", err)
	}

	// Default values
	UUID := generateUUID()
	viper.SetDefault("UUID", UUID)
	viper.SetDefault("DisplayName", "Host-"+UUID)
	viper.SetDefault("DataDir", defaultDataDir)
	viper.SetDefault("RemoteInstances", map[string]RemoteInstance{})
	viper.SetDefault("Sensors", map[string]Sensor{})

	// Try to parse possible existing yaml file or create it
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			logger.Println("Creating empty configuration file")
			if err := viper.WriteConfig(); err != nil {
				logger.Fatal(err)
			}
		} else {
			logger.Fatal("Error reading the configuration:", err)
		}
	}

	// Read everything into the LocalConfig struct
	config.UUID = viper.GetString("UUID")
	config.DisplayName = viper.GetString("DisplayName")
	config.DataDir = viper.GetString("DataDir")

	if err := viper.UnmarshalKey("RemoteInstances", &config.RemoteInstances); err != nil {
		logger.Fatal(err)
	}
	if err := viper.UnmarshalKey("Sensors", &config.Sensors); err != nil {
		logger.Fatal(err)
	}

	// Set viper instance for Sensors access
	config.config = viper.GetViper()

}

func initSensors() {
	// Read measurements
	for _, sensor := range config.Sensors {
		sensor.measurements = viper.New()
		sensor.measurements.SetDefault("measurements", []SensorMeasurement{})
		sensor.measurements.SetConfigFile(config.DataDir + sensor.UUID + ".json")
		if err := sensor.measurements.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				// In case measurements file did not exist, simply create it
				logger.Printf("Info: Creating empty measurements file for sensor %s [%s]", sensor.UUID, sensor.DisplayName)
				if err := sensor.measurements.WriteConfig(); err != nil {
					logger.Fatal("Error creating the measurements file:", err)
				}
			} else {
				// Otherwise
				logger.Printf("Error initializing sensor %s [%s]\n", sensor.UUID, sensor.DisplayName)
				logger.Println(err)
				logger.Printf("Skipping initialization of sensor\n")
				continue
			}
		}

		go sensor.enableMeasurements()
	}

	// Wait forever, in an efficient manner
	select {}
}
