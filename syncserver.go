package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

func handleConn(conn net.Conn) {
	tlsConn := conn.(*tls.Conn)

	buf := make([]byte, 4096)
	tlsConn.Read(buf)

	fmt.Println("Received: ", string(buf))
}

func startSyncServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "{\"msg\": \"Hello, TLS!\"}")
	})

	err := http.ListenAndServeTLS(":8333", local.DataDir+"cert.pem", local.DataDir+"key.pem", nil)
	if err != nil {
		panic(err)
	}
}

func prepareForRemotes() {
	generateTLSCerts()
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
