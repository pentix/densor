package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func handleConn(conn net.Conn) {
	// Socket initialization
	tlsConn := conn.(*tls.Conn)
	tlsConn.SetDeadline(time.Time{})
	defer tlsConn.Close()

	// Protocol data for this connection
	var originUUID string
	var sha256Sum string

	//enc := json.NewEncoder(tlsConn)
	dec := json.NewDecoder(tlsConn)

	// Start TLS Connection
	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("Error on TLS handshake:", err)
	}

	// Start decoding the received requests
	var req Request
	err := dec.Decode(&req)

	if err != nil {
		fmt.Println("Error decoding stuff:", err)
	}

	// First request should be an identifying connection attempt
	if req.RequestType != RequestTypeConnectionAttempt {
		logger.Printf("First request of %s was not a valid connection attempt (Type: %d", req.OriginUUID, req.RequestType)
		return
	}

	originUUID = req.OriginUUID
	sha256Sum = SHA256FromTLSCert(tlsConn.ConnectionState().PeerCertificates[0])
	fmt.Println("SHA256 from Client:", sha256Sum)

	// Now verify the identity
	if !matchesAuthorizedKey(originUUID, sha256Sum) {
		return
	}

	// Respond with ACK
	//tlsConn.

	/**
	At this point we are messages are the connection is established,
	encrypted, authenticated and the protocol is initialized.
	*/

}

func startSyncServer() {
	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{local.keyPair}, ClientAuth: tls.RequireAnyClientCert}
	listener, err := tls.Listen("tcp", ":8333", tlsConfig)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go handleConn(conn)
	}
}

func prepareForRemotes() {
	//generateTLSCerts()
}
