package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/websocket"
)

func StartWebInterface() {
	http.HandleFunc("/", WebUIRoot)
	http.HandleFunc("/api", WebAPI)

	http.ListenAndServeTLS("0.0.0.0:8334", local.DataDir+"cert.pem", local.DataDir+"key.pem", nil)
}

func WebUIRoot(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("web/index.html")
	if err != nil {
		logger.Printf("Error: Web Interface: Could not load template:", err)
	}

	t.Execute(w, local)
}

const (
	WebAPIRequestSensorList         = 1
	WebAPIAnswerSensorList          = 2
	WebAPIRequestRemoteInstanceList = 3
	WebAPIAnswerRemoteInstanceList  = 4
)

type WebAPIRequest struct {
	RequestType int
	Data        string
}

func WebAPI(w http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.Println("Error: Web Interface:", err)
		return
	}

	logger.Println("Info: Web Interface: Opening API Connection")
	defer logger.Println("Info: Web Interface: Closed API Connection")
	defer c.Close()

	apiReq := new(WebAPIRequest)
	for {
		err := c.ReadJSON(&apiReq)
		if err != nil {
			logger.Println("Info: Web Interface: Could not decode request:", err)
			continue
		}

		switch apiReq.RequestType {
		case WebAPIRequestSensorList:
			sensorData := make([]map[string]interface{}, 0)

			for _, s := range local.sensors {
				sensorData = append(sensorData, map[string]interface{}{
					"UUID":             s.UUID,
					"InstanceUUID":     local.UUID,
					"DisplayName":      s.DisplayName,
					"LastUpdateTime":   s.lastUpdateTimestamp(),
					"LastUpdateStatus": s.lastUpdateStatus(),
				})
			}

			for _, r := range local.RemoteInstances {
				for _, s := range r.sensors {
					sensorData = append(sensorData, map[string]interface{}{
						"UUID":             s.UUID,
						"InstanceUUID":     r.UUID,
						"DisplayName":      s.DisplayName,
						"LastUpdateTime":   s.lastUpdateTimestamp(),
						"LastUpdateStatus": s.lastUpdateStatus(),
					})
				}
			}

			c.WriteJSON(map[string]interface{}{
				"RequestType": WebAPIAnswerSensorList,
				"Sensors":     sensorData,
			})
			break

		case WebAPIRequestRemoteInstanceList:
			remotesData := make([]map[string]interface{}, 0)
			for _, r := range local.RemoteInstances {
				remotesData = append(remotesData, map[string]interface{}{
					"UUID":        r.UUID,
					"DisplayName": r.DisplayName,
					"Sensors":     r.SensorUUIDs,
					"Connected":   r.connected,
				})
			}

			c.WriteJSON(map[string]interface{}{
				"RequestType":     WebAPIAnswerRemoteInstanceList,
				"RemoteInstances": remotesData,
			})
			break

		}
	}
}
