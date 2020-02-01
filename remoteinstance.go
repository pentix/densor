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
}

func (r *RemoteInstance) Connect() {
	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{local.keyPair}, ClientAuth: tls.RequireAnyClientCert}
	tlsConn, err := tls.Dial("tcp", r.RemoteAddress, tlsConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	r.tlsConn = tlsConn
	r.connected = true
	r.tlsConn.SetDeadline(time.Time{})

	fmt.Println("TLS Cert from Server:", SHA256FromTLSCert(r.tlsConn.ConnectionState().PeerCertificates[0]))
	r.SendRequest(Request{
		RequestType: RequestTypeConnectionAttempt,
		OriginUUID:  local.UUID,
		Data: map[string]string{
			"Test": "1234",
		},
	})

}

func (r *RemoteInstance) SendRequest(req Request) {
	if !r.connected {
		fmt.Println("Error: Not connected to remote instance. (Yet trying to send a request)")
		return
	}

	encoder := json.NewEncoder(r.tlsConn)

	if err := encoder.Encode(req); err != nil {
		fmt.Println("Error encoding request:", err)
	}
}

func connectToRemoteInstances() {
	remote := RemoteInstance{
		UUID:          "Remotey",
		DisplayName:   "Remotee",
		RemoteAddress: "localhost:8333",
	}

	remote.Connect()
}
