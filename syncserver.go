package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"time"
)

func handleConn(conn net.Conn) {
	tlsConn := conn.(*tls.Conn)
	tlsConn.SetDeadline(time.Time{})

	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("Error on TLS handshake:", err)
	}

	certUsed := tlsConn.ConnectionState().PeerCertificates[0]
	sha256sum := sha256.Sum256(certUsed.Raw)
	fmt.Println(hex.EncodeToString(sha256sum[:]))

	var req Request
	dec := json.NewDecoder(tlsConn)

	err := dec.Decode(&req)

	if err != nil {
		fmt.Println("Error decoding stuff:", err)
	}

	fmt.Println("Received: ", req)
}

func startSyncServer() {
	keyPair, err := tls.LoadX509KeyPair(local.DataDir+"cert.pem", local.DataDir+"key.pem")

	tlsConfig := &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{keyPair}, ClientAuth: tls.RequireAnyClientCert}
	listener, err := tls.Listen("tcp", ":8333", tlsConfig)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Accepted connection")
		go handleConn(conn)
	}
}

func prepareForRemotes() {
	//generateTLSCerts()
}

func generateTLSCerts() error {
	key, err := rsa.GenerateKey(rand.Reader, 768)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyFile, err := os.OpenFile(local.DataDir+"key.pem", os.O_CREATE|os.O_WRONLY, 0700)
	certFile, err := os.OpenFile(local.DataDir+"cert.pem", os.O_CREATE|os.O_WRONLY, 0700)

	err = pem.Encode(io.Writer(keyFile), &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		return err
	}

	err = pem.Encode(io.Writer(certFile), &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return err
	}

	keyFile.Close()
	certFile.Close()

	return err
}
