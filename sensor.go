package main

import (
	"encoding/json"
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
	NextMeasurement int
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

		measurement := SensorMeasurement{
			SensorUUID:    s.UUID,
			MeasurementId: s.NextMeasurement,
			Timestamp:     time.Now().String(),
			Error:         false,
			Data:          make(map[string]interface{}),
		}

		// Collect sensor data
		output, err := command.Output()
		if err != nil {
			return measurement, err
		}

		// If execution was successful, add the output
		measurement.Data["Output"] = string(output)
		return measurement, err

	default:
		return SensorMeasurement{}, fmt.Errorf("unknown sensor type: %d", s.Type)
	}
}

func (s *Sensor) addMeasurement(measurement *SensorMeasurement, bulkInsert bool) {
	// Update internal struct if there is an actual measurement to be inserted
	if measurement != nil {
		s.Measurements = append(s.Measurements, *measurement)
		s.NextMeasurement += 1
	}

	// If we bulkInsert a nil measurement this means we should write to viper/disk!
	// If bulkInsert is false, we save every single entry directly to viper/disk
	if !bulkInsert || (bulkInsert && measurement == nil) {
		s.sensorFile.Set("Sensor", s) // todo mutex
		if err := s.sensorFile.WriteConfig(); err != nil {
			logger.Println("Error: Writing measurement to disk failed:", err)
		}
	}
}

func (s *Sensor) lastUpdateStatus() bool {
	if s.Measurements[len(s.Measurements)-1].Error { // Todo take period into account (update should be new)
		return false
	}
	return true
}

func (s *Sensor) lastUpdateTimestamp() string {
	return s.Measurements[len(s.Measurements)-1].Timestamp[0:19]
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
			measurement.Error = true
			measurement.Data["ErrorMessage"] = fmt.Sprint(err)
		}

		s.addMeasurement(&measurement, false)

		// Directly encode the measurement as an update and broadcast it to all connected remoteInstances
		update := SensorUpdateList{
			SensorMeasurements: map[string][]SensorMeasurement{s.UUID: []SensorMeasurement{
				measurement,
			}},
		}

		enc, err := json.Marshal(update)
		if err != nil {
			logger.Println("Error: Sensor: Failed when trying to create update broadcast")
		}

		BroadcastRequest(&Request{
			RequestType: RequestTypeAnswerSensorMeasurements,
			OriginUUID:  local.UUID,
			Data: map[string]string{
				"collectedUpdates": string(enc),
			},
		})

		time.Sleep(period)
	}
}
