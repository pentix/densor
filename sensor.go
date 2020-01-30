package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/viper"
)

const (
	SensorTypeBasic = 0
)

type Sensor struct {
	UUID            string
	DisplayName     string
	Type            int
	NextMeasurement int64
	Settings        map[string]interface{}

	measurements *viper.Viper
}

func (s *Sensor) incrementNextMeasurement() {
	compositeKey := "Sensors." + s.UUID + ".NextMeasurement"
	newVal := config.config.GetInt64(compositeKey) + 1
	config.config.Set(compositeKey, newVal)
}

func (s *Sensor) settingsString(key string) string {
	compositeKey := "Sensors." + s.UUID + ".settings." + key
	logger.Println("Asking for", compositeKey)
	return config.config.GetString(compositeKey)
}

func (s *Sensor) settingsStringSlice(key string) []string {
	compositeKey := "Sensors." + s.UUID + ".settings." + key
	return config.config.GetStringSlice(compositeKey)
}

func (s *Sensor) settingsStringMap(key string) map[string]string {
	compositeKey := "Sensors." + s.UUID + ".settings." + key
	return config.config.GetStringMapString(compositeKey)
}

func (s *Sensor) sense() (SensorMeasurement, error) {
	switch s.Type {
	case SensorTypeBasic:
		executable := s.settingsString("executable")
		args := s.settingsStringSlice("args")

		command := exec.Command(executable, args...)

		// Collect sensor data
		sensorData := make(map[string]interface{})
		output, err := command.Output()
		if err != nil {
			return SensorMeasurement{}, err
		}

		sensorData["output"] = string(output)

		measurement := SensorMeasurement{
			SensorUUID:    s.UUID,
			MeasurementId: s.NextMeasurement,
			Timestamp:     time.Now(),
			Error:         false,
			Data:          sensorData,
		}

		logger.Println("Measurement completed:", measurement)
		return measurement, err

	default:
		return SensorMeasurement{}, fmt.Errorf("unknown sensor type: %d", s.Type)
	}
}

// Will run in an infinite loop, supposed to run as own goroutine
func (s *Sensor) enableMeasurements() {
	for {
		measurement, err := s.sense()
		if err != nil {
			logger.Printf("Error in sensor %s [%s]: %s", s.UUID, s.DisplayName, err)
		} else {
			soFar := s.measurements.GetSllice("measurements")
			s.measurements.Set("measurements", s.measurements.Get("measurements").append(measurement))
			s.incrementNextMeasurement()
			if err := s.measurements.WriteConfig(); err != nil {
				logger.Printf("Error in sensor %s [%s] saving measurement to disk: %s", s.UUID, s.DisplayName, err)
			}

			if err := config.config.WriteConfig(); err != nil {
				logger.Printf("Error in sensor %s [%s] updating NextMeasurement: %s", s.UUID, s.DisplayName, err)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
