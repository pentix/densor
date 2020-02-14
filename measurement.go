package main

type SensorMeasurement struct {
	SensorUUID    string
	MeasurementId int
	Timestamp     string // Viper issue: https://github.com/spf13/viper/issues/496
	Error         bool
	Data          map[string]interface{}
}
