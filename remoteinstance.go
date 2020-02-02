package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string

	tlsConn   *tls.Conn
	connected bool

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
				r.connected = false

				return
			}

			logger.Println("Error decoding request:", err)
			continue
		}

		// Connection is already established and acknowledged, i.e.
		// no RequestTypeConnectionAttempt and no RequestTypeConnectionACK
		switch req.RequestType {

		default:
			logger.Println("Error: Unexpected request type:", req.RequestType)
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
	r.SendRequest(Request{
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

	logger.Println("yeaaa boiii", r)
	debug()
	logger.Println("oof")

	return true
}

func (r *RemoteInstance) Disconnect() {
	r.tlsConn.Close()
	r.connected = false

	logger.Println("Info: RemoteInstance: Disconnected from", r.UUID)
}

func (r *RemoteInstance) SendRequest(req Request) {
	if err := r.enc.Encode(req); err != nil {
		fmt.Println("Error encoding request:", err)
	}
}

func connectToRemoteInstances() {
	logger.Println("Info: Trying to connect to remote instances")

	// todo: mutex
	for i, _ := range local.RemoteInstances {
		if local.RemoteInstances[i].Connect() {
			local.RemoteInstances[i].HandleIncomingRequests()
		} else {
			logger.Println("Error: Could not connect to", local.RemoteInstances[i].UUID)
		}
	}
}
