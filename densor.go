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

	logger.Println("Instance UUID:              ", config.UUID)
	logger.Println("Instance DisplayName:       ", config.DisplayName)

	initSensors()
}
