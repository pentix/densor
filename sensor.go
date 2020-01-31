package main

import (
	"fmt"
	"os/exec"
	"reflect"
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
	return fmt.Sprint(s.Settings[key])
}

func (s *Sensor) settingsStringSlice(key string) []string {
	if reflect.TypeOf(s.Settings[key]).Kind() != reflect.Slice {
		logger.Println("Error parsing string slice from settings of sensor", s.UUID)
		return nil
	}

	v := reflect.ValueOf(s.Settings[key])
	l := v.Len()

	stringSlice := make([]string, l)
	for i := 0; i < l; i++ {
		stringSlice[i] = fmt.Sprint(v.Index(i))
	}

	return stringSlice
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

		sensorData["Output"] = string(output)

		measurement := SensorMeasurement{
			SensorUUID:    s.UUID,
			MeasurementId: s.NextMeasurement,
			Timestamp:     time.Now().String(),
			Error:         false,
			Data:          sensorData,
		}

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
	period, err := time.ParseDuration(s.settingsString("period"))
	if err != nil {
		logger.Printf("Error parsing period of sensor %s: %s", s.UUID, err)
		logger.Println("Using 10 minutes as default period")
		period = 10 * time.Minute
	}

	logger.Printf("Enabled measurements for %s [%s] @ %v", s.UUID, s.DisplayName, period)

	for {
		measurement, err := s.sense()
		if err != nil {
			logger.Printf("Error in sensor %s [%s]: %s", s.UUID, s.DisplayName, err)
		} else {
			s.addMeasurement(measurement)
		}

		time.Sleep(period)
	}
}
