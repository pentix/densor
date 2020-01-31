package main

import (
	"fmt"
	"log"
	"os"
	"time"
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

	go startSensors()

	// if  --dashboard  show dashboard
	for {
		showDashboard()
		time.Sleep(1 * time.Second)
	}

	// else block!
	// select {}
}

func showDashboard() {
	fmt.Println("\033[2J")
	fmt.Println("------------------------------------------------------------------------------------------------------------")
	fmt.Println("\u001b[1mInstance                                 Sensor                         Status         Last Update\u001b[0m")
	fmt.Println("------------------------------------------------------------------------------------------------------------")

	// Local sensors first
	for _, sensor := range local.sensors {
		fmt.Printf("%-35s      %-22s         [ OK ]         %s\n", local.DisplayName, sensor.DisplayName, sensor.lastUpdateTimestamp())
	}

	fmt.Println("------------------------------------------------------------------------------------------------------------")
}
