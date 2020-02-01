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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultDataDir := homeDir + "/.densor/"
	if err := os.Mkdir(defaultDataDir, 0755); err != nil && !os.IsExist(err) {
		panic(fmt.Errorf("Could not create default data directory: %s", err))
	}

	logfile, err := os.OpenFile(defaultDataDir+"/densor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	logger = log.New(logfile, "", log.LstdFlags)
	logger.Println("Starting densor...")
	readConfig()

	logger.Println("-----------------------------------------------------------------------------")
	logger.Println("Number of remote instances: ", len(local.RemoteInstanceUUIDs))
	logger.Println("Number of local sensors:    ", len(local.SensorsUUIDs))
	logger.Println("Data Directory:             ", local.DataDir)
	logger.Println("Instance UUID:              ", local.UUID)
	logger.Println("Instance DisplayName:       ", local.DisplayName)
	logger.Println("-----------------------------------------------------------------------------")

	loadTLSCerts()

	go startSyncServer()
	go connectToRemoteInstances()
	go startSensors()

	// if  --dashboard  show dashboard
	for {
		//showDashboard()
		time.Sleep(1 * time.Second)
	}

	// else block!
	// select {}
}

func showDashboard() {
	fmt.Println("\033[2J\033[H")
	fmt.Println("------------------------------------------------------------------------------------------------------------")
	fmt.Println("\033[1mInstance                                 Sensor                          Status     Last Update\033[0m")
	fmt.Println("------------------------------------------------------------------------------------------------------------")

	// Local sensors first
	for _, sensor := range local.sensors {
		statusText, status := sensor.lastUpdateStatus()
		var colorCode string
		if status {
			colorCode = "\033[32m"
		} else {
			colorCode = "\033[31m"
		}

		fmt.Printf("%s%-35s      %-22s          %-4s       %s\033[0m\n",
			colorCode,
			local.DisplayName,
			sensor.DisplayName,
			statusText,
			sensor.lastUpdateTimestamp())
	}

	fmt.Println("------------------------------------------------------------------------------------------------------------")
}
