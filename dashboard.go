package main

import "fmt"

func showDashboard() {
	fmt.Println("\033[2J\033[H")
	fmt.Println("------------------------------------------------------------------------------------------------------------")
	fmt.Println("\033[1mInstance                                 Sensor                          Status     Last Update\033[0m")
	fmt.Println("------------------------------------------------------------------------------------------------------------")

	// Local sensors first
	for _, sensor := range local.sensors {
		printSensorStatus(local.DisplayName, sensor)
	}

	// Then ordered by remote instances
	for _, remote := range local.RemoteInstances {

		// todo: Design decision
		/*
			// Only try to read from connected instances
			if !remote.connected {
				continue
			}
		*/

		for _, sensor := range remote.sensors {
			printSensorStatus(remote.DisplayName, sensor)
		}
	}

	fmt.Println("------------------------------------------------------------------------------------------------------------")

	// Remote instances
	fmt.Println("\n\n")
	fmt.Println("------------------------------------------------------------------------------------------------------------")
	fmt.Println("\033[1mRemote Instance                          Status       \033[0m")
	fmt.Println("------------------------------------------------------------------------------------------------------------")

	for _, remote := range local.RemoteInstances {
		printRemoteInstanceStatus(&remote)
	}

	fmt.Println("------------------------------------------------------------------------------------------------------------")

}

func printSensorStatus(instanceDisplayName string, sensor *Sensor) {
	var statusText string
	var colorCode string
	var lastUpdateTimestamp string

	status := sensor.lastUpdateStatus()
	switch status {
	case SensorStatusOK:
		statusText = "OK"
		colorCode = "\033[32m"
		break
	case SensorStatusFAIL:
		statusText = "FAIL"
		colorCode = "\033[31m"
		break
	case SensorStatusOLD:
		statusText = "OLD"
		colorCode = "\033[31m"
		break
	case SensorStatusSYNC:
		colorCode = "\033[93m"
		statusText = "SYNC"
	}

	lastUpdateTimestamp = sensor.lastUpdateTimestamp()
	fmt.Printf("%s%-35s      %-22s          %-4s       %s\033[0m\n",
		colorCode,
		instanceDisplayName,
		sensor.DisplayName,
		statusText,
		lastUpdateTimestamp)
}

func printRemoteInstanceStatus(remote *RemoteInstance) {
	var statusText string
	var colorCode string

	if remote.connected {
		statusText = "Connected"
		colorCode = "\033[32m"
	} else {
		statusText = "Not Connected"
		colorCode = "\033[31m"
	}

	fmt.Printf("%s%-35s      %-22s\033[0m\n",
		colorCode,
		remote.DisplayName,
		statusText)
}
