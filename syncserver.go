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
		logger.Printf("First request of %s was not a valid connection attempt (Type: %d)", req.OriginUUID, req.RequestType)
		return
	}

	sha256Sum := SHA256FromTLSCert(tlsConn.ConnectionState().PeerCertificates[0])

	// Now verify the identity
	if !matchesAuthorizedKey(req.OriginUUID, sha256Sum) {
		logger.Println("Info: SyncServer: Rejecting connection due to unauthorized key")
		return
	}

	// Todo: Mutex
	// Don't allow multiple connections with the same instance
	var remote *RemoteInstance
	for i, _ := range local.RemoteInstances {
		if local.RemoteInstances[i].UUID == req.OriginUUID {
			if local.RemoteInstances[i].connected {
				// If we are already connected to this instance, we reject this connection
				enc.Encode(Request{
					RequestType: RequestTypeConnectionNACK,
					OriginUUID:  local.UUID,
					Data:        map[string]string{},
				})

				tlsConn.Close()
				return
			}

			// If we weren't connected before, complete the existing instance by adding
			// the required instances / data structures
			local.RemoteInstances[i].tlsConn = tlsConn
			local.RemoteInstances[i].connected = true
			local.RemoteInstances[i].enc = enc
			local.RemoteInstances[i].dec = dec
			local.RemoteInstances[i].nextRequests = make(chan *Request, 2048)

			remote = &local.RemoteInstances[i]
		}
	}

	// If we didn't know this connection before, add it to our list
	// (first time connection, UUID and Cert should already be in the authorizedKey file)
	if remote == nil {

		hostAddr, _, _ := net.SplitHostPort(tlsConn.RemoteAddr().String())

		// Create a remoteInstance for this connection
		remoteToAppend := RemoteInstance{
			UUID:          req.OriginUUID,
			DisplayName:   req.Data["DisplayName"],
			RemoteAddress: hostAddr + ":8333",
			SensorUUIDs:   []string{},
			sensors:       []*Sensor{},
			tlsConn:       tlsConn,
			connected:     true,
			enc:           enc,
			dec:           dec,
			nextRequests:  make(chan *Request, 2048),
		}

		local.RemoteInstances = append(local.RemoteInstances, remoteToAppend)
		remote = &local.RemoteInstances[len(local.RemoteInstances)-1]

		// todo: mutex?
		local.config.Set("RemoteInstances", local.RemoteInstances)
		local.config.WriteConfig()
	}

	// Respond with ACK
	remote.SendRequest(&Request{
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

	// Notify UI
	WebAPIBroadcastRemoteInstances()

	go remote.MultiplexRequests()
	go remote.GeneratePeriodicRequests()

	// Ask for sensor updates
	remote.nextRequests <- &Request{
		RequestType: RequestTypeGetSensorList,
		OriginUUID:  local.UUID,
		Data:        map[string]string{},
	}

	remote.HandleIncomingRequests()
}

func startSyncServer() {
	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{local.keyPair}, ClientAuth: tls.RequireAnyClientCert}
	listener, err := tls.Listen("tcp", ":8333", tlsConfig)
	if err != nil {
		panic(err)
	}

	logger.Println("Info: SyncServer: Listening on", listener.Addr().String(), "for incoming requests")

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
