package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"
)

func loadTLSCerts() {
	keyPair, err := tls.LoadX509KeyPair(local.DataDir+"cert.pem", local.DataDir+"key.pem")
	if err != nil {
		if os.IsNotExist(err) {
			// Keys don't exist --> Generate now!
			fmt.Println("No TLS Certificate found. Generating now...")
			if err := generateTLSCerts(); err != nil {
				fmt.Println("Could not generate TLS Certificates:", err)
				logger.Fatal(err)
			}

			// Retry loading exactly once!
			keyPair, err = tls.LoadX509KeyPair(local.DataDir+"cert.pem", local.DataDir+"key.pem")
			if err != nil {
				logger.Fatal(err)
			}

		} else {
			// Something else went wrong!
			fmt.Println("Error reading TLS Certificates:", err)
			logger.Fatal(err)
		}
	}

	local.keyPair = keyPair
	logger.Println("Loaded TLS Certificate from files")
}

func generateTLSCerts() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Serials should not collide, otherwise Browsers might complain if multiple instances try to access
	// the Web UI and a cert using the same serial number is already registered locally. (It is however not relevant to security)
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<63-1))

	template := x509.Certificate{SerialNumber: serial, NotBefore: time.Now(), NotAfter: time.Now().Add(10 * 365 * 24 * time.Hour)}
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

func SHA256FromTLSCert(cert *x509.Certificate) string {
	sha256sum := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sha256sum[:])
}

func matchesAuthorizedKey(instanceUUID, sha256Sum string) bool {
	if local.authorizedKeys.IsSet(instanceUUID) {
		if local.authorizedKeys.Get(instanceUUID) == sha256Sum {
			return true
		} else {
			logger.Println("Error: Authorized Keys: Instance", instanceUUID, "provided wrong certificate!")
			return false
		}
	}

	logger.Println("Error: Authorized Keys: No entry for instance", instanceUUID)
	return false
}
