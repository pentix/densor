package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	// Todo: defer close and
	// Todo: cleanup in local.remoteInstances

	fmt.Println("Info: Handling requests of", r.UUID)
	for {
		var req Request
		if err := r.dec.Decode(&req); err != nil {
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
	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{local.keyPair}, ClientAuth: tls.RequireAnyClientCert}
	tlsConn, err := tls.Dial("tcp", r.RemoteAddress, tlsConfig)
	if err != nil {
		fmt.Println(err)
		return false
	}

	tlsConn.SetDeadline(time.Time{})

	r.tlsConn = tlsConn
	r.enc = json.NewEncoder(r.tlsConn)
	r.dec = json.NewDecoder(r.tlsConn)

	// Verify identity
	sha256Sum := SHA256FromTLSCert(r.tlsConn.ConnectionState().PeerCertificates[0])
	fmt.Println("TLS Cert from Server:", sha256Sum)
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
		logger.Println("Did not receive acknowledgement from host", ack.OriginUUID)
		return false
	}

	logger.Println("Connected to", ack.OriginUUID)
	r.connected = true
	return true
}

func (r *RemoteInstance) SendRequest(req Request) {
	if !r.connected {
		fmt.Println("Error: Not connected to remote instance. (Yet trying to send a request)")
		return
	}

	if err := r.enc.Encode(req); err != nil {
		fmt.Println("Error encoding request:", err)
	}
}

// Demo only
func connectToRemoteInstances() {
	remote := RemoteInstance{
		UUID:          "Remotey",
		DisplayName:   "Remotee",
		RemoteAddress: "localhost:8333",
	}

	// Todo: mutex
	local.remoteInstances = append(local.remoteInstances, &remote)

	if remote.Connect() {
		// avoid demo loop
		//remote.HandleIncomingRequests()
	}
}
