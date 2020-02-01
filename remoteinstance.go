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
	//err := generateTLSCerts()

	keyPair, err := tls.LoadX509KeyPair(local.DataDir+"cert.pem", local.DataDir+"key.pem")

	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{keyPair}, ClientAuth: tls.RequireAnyClientCert}
	r.tlsConn, err = tls.Dial("tcp", r.RemoteAddress, tlsConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	r.connected = true
	r.tlsConn.SetDeadline(time.Time{})

	fmt.Println("TLS Cert from Server:", SHA256FromTLSConn(r.tlsConn))

	fmt.Println("Sending request")
	r.SendRequest(Request{
		RequestType: RequestTypeConnectionAttempt,
		OriginUUID:  local.UUID,
		Data: map[string]string{
			"Test": "1234",
		},
	})
	fmt.Println("Request sent :)")

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
