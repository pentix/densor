package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

type RemoteInstance struct {
	UUID          string
	DisplayName   string
	RemoteAddress string

	tlsClient http.Client
}

func (r *RemoteInstance) Connect() {
	err := generateTLSCerts()

	tlsConfig := tls.Config{InsecureSkipVerify: true}
	transport := http.Transport{TLSClientConfig: &tlsConfig}
	r.tlsClient = http.Client{Transport: &transport}

	resp, err := r.tlsClient.Get(r.RemoteAddress)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	values := make(map[string]interface{})
	err = dec.Decode(&values)

	fmt.Println("Err:", err)
	fmt.Println(values)

	r.SendRequest(Request{
		RequestType: RequestTypeConnectionAttempt,
		OriginUUID:  local.UUID,
		Data: map[string]string{
			"Test": "1234",
		},
	})

}

func (r *RemoteInstance) SendRequest(req Request) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)

	if err := encoder.Encode(req); err != nil {
		fmt.Println("Error encoding request:", err)
	}
	r.tlsClient.Post("/", "text/json", buffer)
}
