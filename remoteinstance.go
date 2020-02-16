package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"time"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string
	SensorUUIDs   []string

	sensors      []*Sensor
	tlsConn      *tls.Conn
	connected    bool
	nextRequests chan *Request

	enc *json.Encoder
	dec *json.Decoder
}

func (r *RemoteInstance) HandleIncomingRequests() {
	logger.Printf("Info: RemoteInstance: Connected to %s. Now handling requests\n", r.UUID)
	defer r.Disconnect()

	for {
		var req Request
		if err := r.dec.Decode(&req); err != nil {
			if err == io.EOF {
				logger.Printf("Info: RemoteInstance: %s closed connection. No longer connected.", r.UUID)

				return
			}

			logger.Println("Error decoding request:", err)
			return
		}

		// Connection is already established and acknowledged, i.e.
		// no RequestTypeConnectionAttempt and no RequestTypeConnectionACK
	RequestDestinction:
		switch req.RequestType {
		case RequestTypeGetSensorList:
			logger.Printf("Info: RemoteInstance: %s asks for the sensor list", req.OriginUUID)

			entries := make([]SensorListEntry, 0)
			for _, s := range local.sensors {
				entries = append(entries, SensorListEntry{
					UUID:            s.UUID,
					DisplayName:     s.DisplayName,
					NextMeasurement: s.NextMeasurement,
				})
			}

			// Encode the entries manually
			enc, err := json.Marshal(entries)
			if err != nil {
				logger.Printf("Error: RemoteInstance: Could not collect sensor data requested by %s", req.OriginUUID)
				break RequestDestinction
			}

			r.nextRequests <- &Request{
				RequestType: RequestTypeAnswerSensorList,
				OriginUUID:  local.UUID,
				Data: map[string]string{
					"entries": string(enc),
				},
			}

			break

		case RequestTypeAnswerSensorList:
			logger.Printf("Info: RemoteInstance: %s answered with its sensors", req.OriginUUID)

			entries := make([]SensorListEntry, 0)
			err := json.Unmarshal([]byte(req.Data["entries"]), &entries)

			if err != nil {
				logger.Println("Error: RemoteInstance: Could not decode sensor list")
				break RequestDestinction
			}

			requiresUpdate := make([]SensorUpdateRequestEntry, 0)
			for _, entry := range entries {

				// Check if we already know the sensor
				requireAllMeasurements := true
				for _, sensor := range r.sensors {
					if sensor.UUID == entry.UUID {
						requireAllMeasurements = false

						logger.Printf("Sensor:", sensor.UUID)
						logger.Printf("Local DB: next measurement: %d    Remote entry: next measurement: %d", sensor.NextMeasurement, entry.NextMeasurement)

						// Check if we are up to date
						if sensor.NextMeasurement == entry.NextMeasurement {
							break
						}

						// Not up-to-date, but we don't require the all measurements
						requiresUpdate = append(requiresUpdate, SensorUpdateRequestEntry{
							UUID:                  sensor.UUID,
							StartingAtMeasurement: sensor.NextMeasurement,
						})

						break
					}
				}

				if requireAllMeasurements {
					requiresUpdate = append(requiresUpdate, SensorUpdateRequestEntry{
						UUID:                  entry.UUID,
						StartingAtMeasurement: 0,
					})
				}
			}

			logger.Printf("Received %d sensors of which %d require an update or are unknown.", len(entries), len(requiresUpdate))
			logger.Printf("Request: %v", requiresUpdate)

			// Encode the entries manually
			enc, err := json.Marshal(requiresUpdate)
			if err != nil {
				logger.Println("Error: RemoteInstance: Could not encode required updates")
				break RequestDestinction
			}

			// Request update for all sensors in requiresUpdate
			r.nextRequests <- &Request{
				RequestType: RequestTypeGetSensorMeasurements,
				OriginUUID:  local.UUID,
				Data: map[string]string{
					"entries": string(enc),
				},
			}

			break

		case RequestTypeGetSensorMeasurements:
			requestedMeasurements := make([]SensorUpdateRequestEntry, 0)
			if err := json.Unmarshal([]byte(req.Data["entries"]), &requestedMeasurements); err != nil {
				logger.Println("Error: RemoteInstance: Could not decode requested sensor measurements")
				break RequestDestinction
			}

			logger.Printf("Info: RemoteInstance: Remote %s asked for %d sensor updates", req.OriginUUID, len(requestedMeasurements))

			// Init sensor updates collection
			collectedUpdates := SensorUpdateList{
				SensorMeasurements: map[string][]SensorMeasurement{},
			}

			for _, requestedPair := range requestedMeasurements {
				// Collect requested updates
				index := local.GetSensorIndex(requestedPair.UUID)
				if index < 0 || index >= len(local.sensors) {
					logger.Println("Error: RemoteInstance: Invalid request for sensor", requestedPair.UUID)
					break
				}

				if requestedPair.StartingAtMeasurement < 0 ||
					requestedPair.StartingAtMeasurement >= local.sensors[index].NextMeasurement {
					logger.Printf("Error: RemoteInstance: Invalid request for sensor %s measurement %d",
						requestedPair.UUID,
						requestedPair.StartingAtMeasurement)

					break RequestDestinction
				}

				logger.Printf("Info: RemoteInstance: Collected %d measurements from sensor %s starting from %d up to %d",
					len(local.sensors[index].Measurements)-requestedPair.StartingAtMeasurement,
					requestedPair.UUID,
					requestedPair.StartingAtMeasurement,
					len(local.sensors[index].Measurements))
				collectedUpdates.SensorMeasurements[requestedPair.UUID] = local.sensors[index].Measurements[requestedPair.StartingAtMeasurement:]
			}

			enc, err := json.Marshal(collectedUpdates)
			if err != nil {
				logger.Println("Error: RemoteInstance: Could not encode collected updates")
				break
			}

			r.nextRequests <- &Request{
				RequestType: RequestTypeAnswerSensorMeasurements,
				OriginUUID:  local.UUID,
				Data: map[string]string{
					"collectedUpdates": string(enc),
				},
			}

			break

		case RequestTypeAnswerSensorMeasurements:

			collectedUpdates := SensorUpdateList{}
			if err := json.Unmarshal([]byte(req.Data["collectedUpdates"]), &collectedUpdates); err != nil {
				logger.Println("Error: RemoteInstance: Could not decode collected updates")
				break RequestDestinction
			}

			for UUID, measurements := range collectedUpdates.SensorMeasurements {

				// Create the sensor if it appears to be new
				index := r.GetSensorIndex(UUID)
				if index < 0 {
					logger.Printf("Info: RemoteInstance: Learned about sensor %s from remote %s", UUID, r.UUID)
					r.AddSensor(&Sensor{
						UUID:            UUID,
						DisplayName:     "DemoDisplayName", // todo: fix
						Type:            0,                 //todo: fix
						NextMeasurement: 0,
						Settings:        map[string]interface{}{},
						Measurements:    []SensorMeasurement{},
						sensorFile:      nil,
					})

					index = r.GetSensorIndex(UUID)
				}

				// Then start adding the measurements (don't log single live updates)
				if len(measurements) > 1 {
					logger.Printf("Info: RemoteInstance: Received update from sensor %s with %d measurements, starting at %d",
						UUID,
						len(measurements),
						measurements[0].MeasurementId)
				}

				for _, m := range measurements {
					if m.MeasurementId != r.sensors[index].NextMeasurement {
						logger.Println("Error: RemoteInstance: Collected updates are not in correct order")
						break RequestDestinction
					}

					r.sensors[index].addMeasurement(&m, true)
				}

				// "Commit" bulk transaction
				r.sensors[index].addMeasurement(nil, true)
			}

			break

		default:
			logger.Println("Error: RemoteInstance: Unexpected request type:", req.RequestType)
			logger.Println("Error: RemoteInstance: Received request:", req)
		}
	}
}

func (r *RemoteInstance) Connect() bool {
	logger.Println("Info: Connect(): Trying to connect to", r.UUID)

	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{local.keyPair}, ClientAuth: tls.RequireAnyClientCert}
	tlsConn, err := tls.Dial("tcp", r.RemoteAddress, tlsConfig)
	if err != nil {
		logger.Println("Info: Could not connect to", r.UUID, ":", err)
		return false
	}

	tlsConn.SetDeadline(time.Time{})

	r.tlsConn = tlsConn
	r.enc = json.NewEncoder(r.tlsConn)
	r.dec = json.NewDecoder(r.tlsConn)

	// Verify identity
	sha256Sum := SHA256FromTLSCert(r.tlsConn.ConnectionState().PeerCertificates[0])
	if !matchesAuthorizedKey(r.UUID, sha256Sum) {
		return false
	}

	// Send ConnectionAttempt
	r.SendRequest(&Request{
		RequestType: RequestTypeConnectionAttempt,
		OriginUUID:  local.UUID,
		Data: map[string]string{
			"DisplayName": local.DisplayName,
		},
	})

	// Wait for ACK
	var ack Request
	r.dec.Decode(&ack)

	if ack.RequestType != RequestTypeConnectionACK {
		logger.Printf("Did not receive acknowledgement from host %s (Received type %d)", ack.OriginUUID, ack.RequestType)
		return false
	}

	logger.Println("Connected to", ack.OriginUUID)
	r.connected = true

	return true
}

func (r *RemoteInstance) Disconnect() {
	r.tlsConn.Close()
	r.connected = false

	logger.Println("Info: RemoteInstance: Disconnected from", r.UUID)
}

func (r *RemoteInstance) SendRequest(req *Request) {
	if err := r.enc.Encode(req); err != nil {
		fmt.Println("Error encoding request:", err)
	}
}

func (r *RemoteInstance) GeneratePeriodicRequests() {
	// todo implement
}

func (r *RemoteInstance) MultiplexRequests() {
	// Sleep until connection is ready and standard handshake collected some sync data
	time.Sleep(500 * time.Millisecond)

	for nextReq := range r.nextRequests {
		r.SendRequest(nextReq)
	}
}

func (r *RemoteInstance) AddSensor(sensor *Sensor) {
	// todo: mutex

	// Create measurements file
	sensor.sensorFile = viper.New()
	sensor.sensorFile.SetConfigFile(local.DataDir + sensor.UUID + ".json")
	sensor.sensorFile.Set("Sensor", sensor)
	sensor.sensorFile.WriteConfig()

	// Add sensor to remote instances and save it
	r.SensorUUIDs = append(r.SensorUUIDs, sensor.UUID)
	r.sensors = append(r.sensors, sensor)
	local.config.Set("RemoteInstances", local.RemoteInstances)
	local.config.WriteConfig()
}

func connectToRemoteInstances() {
	logger.Println("Info: Trying to connect to remote instances")

	// todo: mutex
	for i, _ := range local.RemoteInstances {
		currentRemote := &local.RemoteInstances[i]

		// Prepare multiplexing
		currentRemote.nextRequests = make(chan *Request, 2048)

		if currentRemote.Connect() {
			go currentRemote.HandleIncomingRequests()   // Handle incoming requests
			go currentRemote.MultiplexRequests()        // Enable outgoing message multiplexing
			go currentRemote.GeneratePeriodicRequests() // Activate periodic polling (heartbeats, etc.)

			// First thing to do once we're connected is to ask for the remote instance's sensors
			currentRemote.nextRequests <- &Request{
				RequestType: RequestTypeGetSensorList,
				OriginUUID:  local.UUID,
				Data:        map[string]string{},
			}

		} else {
			logger.Println("Error: Could not connect to", local.RemoteInstances[i].UUID)
		}
	}
}

func (r *RemoteInstance) GetSensorIndex(UUID string) int {
	for i, s := range r.SensorUUIDs {
		if s == UUID {
			return i
		}
	}

	return -1
}
