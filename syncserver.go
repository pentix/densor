package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func handleConn(conn net.Conn) {
	tlsConn := conn.(*tls.Conn)
	tlsConn.SetDeadline(time.Time{})

	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("Error on TLS handshake:", err)
	}

	fmt.Println("SHA256 from Client:", SHA256FromTLSCert(tlsConn.ConnectionState().PeerCertificates[0]))

	var req Request
	dec := json.NewDecoder(tlsConn)

	err := dec.Decode(&req)

	if err != nil {
		fmt.Println("Error decoding stuff:", err)
	}

	fmt.Println("Received: ", req)
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
