package main

const (
	RequestTypeConnectionAttempt = 1
	RequestTypeConnectionACK     = 2
)

type Request struct {
	RequestType int
	OriginUUID  string
	Data        map[string]string
}
