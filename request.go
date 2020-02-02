package main

const (
	RequestTypeConnectionAttempt = 1
	RequestTypeConnectionACK     = 2
	RequestTypeConnectionNACK    = 3
)

type Request struct {
	RequestType int
	OriginUUID  string
	Data        map[string]string
}
