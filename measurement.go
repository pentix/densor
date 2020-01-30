package main

import (
	"time"
)

type SensorMeasurement struct {
	SensorUUID string
	Counter int64
	Timestamp time.Time
	Error bool
	Data map[string]interface{}
}