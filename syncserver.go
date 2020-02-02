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

	enc := json.NewEncoder(tlsConn)
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

	sha256Sum := SHA256FromTLSCert(tlsConn.ConnectionState().PeerCertificates[0])
	fmt.Println("SHA256 from Client:", sha256Sum)

	// Now verify the identity
	if !matchesAuthorizedKey(req.OriginUUID, sha256Sum) {
		return
	}

	// Create a remoteInstance for this connection
	remote := RemoteInstance{
		UUID:          req.OriginUUID,
		DisplayName:   req.Data["DisplayName"],
		RemoteAddress: tlsConn.RemoteAddr().String(),
		tlsConn:       tlsConn,
		connected:     true,
		enc:           enc,
		dec:           dec,
	}

	// Todo: Mutex
	// Don't allow multi-connections
	found := false
	for i, r := range local.remoteInstances {
		if r.UUID == req.OriginUUID {
			if !r.connected {
				// If we know the remote instance, but weren't connected before
				local.remoteInstances[i] = &remote
				found = true
			}

			break
		}
	}

	if !found {
		local.remoteInstances = append(local.remoteInstances, &remote)
	}

	// Respond with ACK
	remote.SendRequest(Request{
		RequestType: RequestTypeConnectionACK,
		OriginUUID:  local.UUID,
		Data:        map[string]string{"DisplayName": local.DisplayName},
	})

	/**
	At this point we are messages are the connection is established,
	encrypted, authenticated and the protocol is initialized.

	Since we established connection, we can now
	wait and handle duplex requests
	*/

	remote.HandleIncomingRequests()
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
