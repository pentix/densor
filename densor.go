package main

import (
	"log"
	"os"
)

var logger *log.Logger
var local LocalInstance

func main() {
	// Read local and start logging
	logger = log.New(os.Stdout, "", log.LstdFlags)
	logger.Println("Starting densor...")
	readConfig()

	logger.Println("-----------------------------------------------------------------------------")
	logger.Println("Number of remote instances: ", len(local.RemoteInstanceUUIDs))
	logger.Println("Number of local sensors:    ", len(local.SensorsUUIDs))
	logger.Println("Data Directory:             ", local.DataDir)
	logger.Println("Instance UUID:              ", local.UUID)
	logger.Println("Instance DisplayName:       ", local.DisplayName)
	logger.Println("-----------------------------------------------------------------------------")

	startSensors()
}
