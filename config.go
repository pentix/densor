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

type LocalInstance struct {
	UUID                string
	DisplayName         string
	DataDir             string
	RemoteInstanceUUIDs []string
	SensorsUUIDs        []string

	remoteInstances []RemoteInstance
	sensors         []Sensor
	config          *viper.Viper
}

func readConfig() {
	// Actual local file
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
	if err := os.Mkdir(defaultDataDir, 0755); err != nil && !os.IsExist(err) {
		logger.Fatal("Could not create default data directory:", err)
	}

	// Default values
	UUID := generateUUID()
	viper.SetDefault("UUID", UUID)
	viper.SetDefault("DisplayName", "Host-"+UUID)
	viper.SetDefault("DataDir", defaultDataDir)
	viper.SetDefault("RemoteInstances", []string{})
	viper.SetDefault("Sensors", []string{})

	// Try to parse possible existing yaml file or create it
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			logger.Println("Creating empty configuration file")
			if err := viper.WriteConfig(); err != nil {
				logger.Fatal("Error creating the configuration:", err)
			}
		} else {
			logger.Fatal("Error reading the configuration:", err)
		}
	}

	// Read everything into the LocalInstance struct
	local.UUID = viper.GetString("UUID")
	local.DisplayName = viper.GetString("DisplayName")
	local.DataDir = viper.GetString("DataDir")

	if err := viper.UnmarshalKey("RemoteInstances", &local.RemoteInstanceUUIDs); err != nil {
		logger.Fatal(err)
	}
	if err := viper.UnmarshalKey("Sensors", &local.SensorsUUIDs); err != nil {
		logger.Fatal(err)
	}

	// Set viper instance for Sensors access
	local.config = viper.GetViper()
}

func startSensors() {
	// Read measurements
	for _, sensorUUID := range local.SensorsUUIDs {

		// Prepare reading sensor data
		reader := viper.New()
		reader.SetDefault("Sensor", Sensor{
			UUID:            sensorUUID,
			DisplayName:     sensorUUID,
			Type:            0,
			NextMeasurement: 0,
			Settings:        map[string]interface{}{"Period": "10m", "Executable": "", "Args": []string{}},
			Measurements:    []SensorMeasurement{},
		})
		reader.SetConfigFile(local.DataDir + sensorUUID + ".json")

		if err := reader.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				logger.Println("Creating default sensor data file. You might want to edit it!")
				if writeErr := reader.WriteConfig(); writeErr != nil {
					logger.Println("Error creating default sensor data file:", writeErr)
				}
			} else {
				logger.Println("Error initializing sensor", sensorUUID, ":", err)
				logger.Println("Skipping initialization of sensor")
				continue
			}
		}

		// Read sensor data into
		var sensor Sensor
		if err := reader.UnmarshalKey("Sensor", &sensor); err != nil {
			logger.Printf("Error: Could not unmarshal sensor %s: %s", sensorUUID, err)
			logger.Println("Skipping sensor!")
			continue
		}

		sensor.sensorFile = reader
		go sensor.enableMeasurements()
	}

	// Wait forever, in an efficient manner
	select {}
}
