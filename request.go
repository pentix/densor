package main

const (
	RequestTypeConnectionAttempt = 1
	RequestTypeConnectionACK     = 2
	RequestTypeConnectionNACK    = 3

	RequestTypeGetSensorList    = 4
	RequestTypeAnswerSensorList = 5

	RequestTypeGetSensorMeasurements    = 6
	RequestTypeAnswerSensorMeasurements = 7
)

type Request struct {
	RequestType int
	OriginUUID  string
	Data        map[string]string
}

type SensorListEntry struct {
	UUID            string
	DisplayName     string
	NextMeasurement int64
}

type SensorUpdateRequestEntry struct {
	UUID                  string
	startingAtMeasurement int64
}
