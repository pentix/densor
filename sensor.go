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
	Measurements    []SensorMeasurement

	sensorFile *viper.Viper
}

func (s *Sensor) settingsString(key string) string {
	compositeKey := "sensor.settings." + key
	logger.Println("Asking for", compositeKey)

	return s.sensorFile.GetString(compositeKey)
}

func (s *Sensor) settingsStringSlice(key string) []string {
	compositeKey := "sensor.settings." + key
	return s.sensorFile.GetStringSlice(compositeKey)
}

func (s *Sensor) settingsStringMap(key string) map[string]string {
	compositeKey := "sensor.settings." + key
	return s.sensorFile.GetStringMapString(compositeKey)
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

func (s *Sensor) addMeasurement(measurement SensorMeasurement) {
	// Internal struct
	s.Measurements = append(s.Measurements, measurement)
	s.NextMeasurement += 1

	// Write to disk
	s.sensorFile.Set("Sensor", s)
	/*s.sensorFile.Set("Sensor.Measurements", s.Measurements)
	s.sensorFile.Set("Sensor.NextMeasurement", s.NextMeasurement)*/
	if err := s.sensorFile.WriteConfig(); err != nil {
		logger.Println("Error: Writing measurement to disk failed:", err)
	}
}

// Will run in an infinite loop, supposed to run as own goroutine
func (s *Sensor) enableMeasurements() {
	logger.Printf("Enabled measurements for %s [%s]", s.UUID, s.DisplayName)

	for {
		measurement, err := s.sense()
		if err != nil {
			logger.Printf("Error in sensor %s [%s]: %s", s.UUID, s.DisplayName, err)
		} else {
			s.addMeasurement(measurement)
		}

		time.Sleep(10 * time.Second)
	}
}
