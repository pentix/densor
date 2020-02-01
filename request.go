package main

const (
	RequestTypeConnectionAttempt = 1
)

type Request struct {
	RequestType int
	OriginUUID  string
	Data        map[string]string
}
