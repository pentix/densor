package main

import (
	"crypto/tls"
	"os"

	"github.com/spf13/viper"
)

type LocalInstance struct {
	UUID            string
	DisplayName     string
	DataDir         string
	SensorsUUIDs    []string
	RemoteInstances []RemoteInstance

	keyPair        tls.Certificate
	sensors        []*Sensor
	config         *viper.Viper
	authorizedKeys *viper.Viper
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

	// This should already exist at this point :)
	defaultDataDir := homeDir + "/.densor/"

	// Default values
	UUID := generateUUID()
	viper.SetDefault("UUID", UUID)
	viper.SetDefault("DisplayName", "Host-"+UUID)
	viper.SetDefault("DataDir", defaultDataDir)
	viper.SetDefault("RemoteInstances", []RemoteInstance{})
	viper.SetDefault("sensors", []string{})
	viper.SetDefault("WebTLSCert", defaultDataDir+"cert.pem")
	viper.SetDefault("WebTLSKey", defaultDataDir+"key.pem")

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

	if err := viper.UnmarshalKey("RemoteInstances", &local.RemoteInstances); err != nil {
		logger.Fatal(err)
	}

	// Prepare sensor files
	for i, _ := range local.RemoteInstances {
		currentRemote := &local.RemoteInstances[i]

		for i, uuid := range currentRemote.SensorUUIDs {
			currentRemote.sensors = append(currentRemote.sensors, &Sensor{})

			currentRemote.sensors[i].sensorFile = viper.New()
			currentRemote.sensors[i].sensorFile.SetConfigFile(local.DataDir + uuid + ".json")
			if err := currentRemote.sensors[i].sensorFile.ReadInConfig(); err != nil {
				logger.Printf("Error: Config: Could read sensor file %s", uuid)
				continue
			}

			if err := currentRemote.sensors[i].sensorFile.UnmarshalKey("Sensor", &currentRemote.sensors[i]); err != nil {
				logger.Printf("Error: Config: Could not unmarshal sensor %s", uuid)
				continue
			}

			logger.Println("Learned about sensor", currentRemote.sensors[i].UUID)
		}
	}

	if err := viper.UnmarshalKey("sensors", &local.SensorsUUIDs); err != nil {
		logger.Fatal(err)
	}

	// Set viper instance for sensors access
	local.config = viper.GetViper()

	// Set viper instance for authorized keys
	local.authorizedKeys = viper.New()
	local.authorizedKeys.SetConfigFile(local.DataDir + "authorizedKeys.json")
	if err := local.authorizedKeys.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			local.authorizedKeys.WriteConfig()
		} else {
			logger.Fatal("Error reading the authorized keys file:", err)
		}
	}
}

func startSensors() {
	// Read measurements
	for _, sensorUUID := range local.SensorsUUIDs {
		logger.Println("Trying to start sensor", sensorUUID)

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

		// Add sensor to the local instance
		local.sensors = append(local.sensors, &sensor)

		// Set sensor data file and start measurements
		sensor.sensorFile = reader
		go sensor.enableMeasurements()
	}

	// Wait forever, in an efficient manner
	select {}
}

func (l *LocalInstance) GetSensorIndex(UUID string) int {
	for i, s := range l.SensorsUUIDs {
		if s == UUID {
			return i
		}
	}

	return -1
}

func (l *LocalInstance) GetRemoteInstanceIndex(UUID string) int {
	for i, r := range l.RemoteInstances {
		if r.UUID == UUID {
			return i
		}
	}

	return -1
}
