package main

import (
	"crypto/x509"
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
	loadTLSCerts()
	localCert, err := x509.ParseCertificate(local.keyPair.Certificate[0])
	if err != nil {
		logger.Fatal("Error parsing local certificate. Exiting.")
	}

	logger.Println("---------------------------------------------------------------------------------------------")
	logger.Println("Number of remote instances: ", len(local.RemoteInstances))
	logger.Println("Number of local sensors:    ", len(local.SensorsUUIDs))
	logger.Println("Data Directory:             ", local.DataDir)
	logger.Println("Instance UUID:              ", local.UUID)
	logger.Println("Instance DisplayName:       ", local.DisplayName)
	logger.Println("Instance TLS Certificate:   ", SHA256FromTLSCert(localCert))
	logger.Println("---------------------------------------------------------------------------------------------")

	go StartWebInterface()
	go startSyncServer()
	go connectToRemoteInstances()
	go startSensors()

	go debug()

	// if  --dashboard  show dashboard
	for {
		//showDashboard()
		time.Sleep(1 * time.Second)
	}

	// else block!
	// select {}
}

func debug() {
	for {
		//	logger.Println(local.RemoteInstances, "\n\n\n")
		time.Sleep(2 * time.Second)
	}
}
