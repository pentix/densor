package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string

	connection net.Conn
}

func (r *RemoteInstance) Connect() {
	err := generateTLSCerts()

	tlsConfig := tls.Config{InsecureSkipVerify: true}
	transport := http.Transport{TLSClientConfig: &tlsConfig}
	tlsClient := http.Client{Transport: &transport}

	resp, err := tlsClient.Get(r.RemoteAddress)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	values := make(map[string]interface{})
	err = dec.Decode(&values)

	fmt.Println("Err:", err)
	fmt.Println(values)

}
