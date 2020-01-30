package main

import (
	"log"
	"os"
)

var logger *log.Logger
var config LocalConfig

func main() {
	// Read config and start logging
	logger = log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("Starting densor...")
	readConfig()

	logger.Println("-----------------------------------------------------------------------------")
	logger.Println("Number of remote instances: ", len(config.RemoteInstances))
	logger.Println("Number of sensors:          ", len(config.Sensors))
	logger.Println("Data Directory:             ", config.DataDir)
	logger.Println("Instance UUID:              ", config.UUID)
	logger.Println("Instance DisplayName:       ", config.DisplayName)
	logger.Println("-----------------------------------------------------------------------------")

	initSensors()
}
