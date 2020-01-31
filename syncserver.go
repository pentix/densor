package main

import (
	"crypto/tls"
	"fmt"
	"net"
)

func handleConn(conn net.Conn) {
	tlsConn := conn.(*tls.Conn)

	buf := make([]byte, 4096)
	tlsConn.Read(buf)

	fmt.Println("Received: ", string(buf))
}

func startSyncServer() {
	listener, err := tls.Listen("tcp", ":8333", generateTLSConfig())
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	go handleConn(conn)
}
