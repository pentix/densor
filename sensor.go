package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/viper"
)

const (
	SensorTypeBasic = 1
)

type Sensor struct {
	UUID string
	DisplayName string
	Type int
	LastMeasurement int64

	measurements *viper.Viper
}

func (s *Sensor) settingsString(key string) string {
	compositeKey := "Sensors." + s.UUID + "." + key
	return config.config.GetString(compositeKey)
}

func (s *Sensor) settingsStringSlice(key string) []string {
	compositeKey := "Sensors." + s.UUID + "." + key
	return config.config.GetStringSlice(compositeKey)
}

func (s *Sensor) settingsStringMap(key string) map[string]string {
	compositeKey := "Sensors." + s.UUID + "." + key
	return config.config.GetStringMapString(compositeKey)
}



func (s *Sensor) sense() (SensorMeasurement, error) {
	switch s.Type {
	case SensorTypeBasic:
		executable := s.settingsString("executable")
		args := s.settingsStringSlice("args")

		command := exec.Command(executable, args...)
		err := command.Run()
		if err != nil {
			return SensorMeasurement{}, err
		}
		
		measurement := SensorMeasurement {
			SensorUUID: s.UUID,
			Counter:    s.LastMeasurement+1,
			Timestamp:  time.Now(),
			Error:      false,
			Data:       nil,
		}

		s.LastMeasurement += 1
		return measurement, err

	default:
		return SensorMeasurement{}, fmt.Errorf("unknown sensor type: %d", s.Type)
	}
}

// Will run in an infinite loop, supposed to run as own goroutine
func (s *Sensor) enableMeasurements() {

}