package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/websocket"
)

func StartWebInterface() {
	http.HandleFunc("/", WebUIRoot)
	http.HandleFunc("/api", WebAPI)

	go WebAPIBroadcast()
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

	WebAPIRegisterSocket(c)
	defer WebAPIUnregisterSocket(c)
	defer c.Close()

	apiReq := new(WebAPIRequest)
	for {
		err := c.ReadJSON(&apiReq) // todo mutex
		if err != nil {
			if err == websocket.ErrCloseSent {
				return
			}

			logger.Println("Info: Web Interface: Could not decode request:", err)
			continue
		}

		switch apiReq.RequestType {
		case WebAPIRequestSensorList:
			sensorData := WebAPICollectSensors()

			c.WriteJSON(map[string]interface{}{
				"RequestType": WebAPIAnswerSensorList,
				"Sensors":     sensorData,
			})
			break

		case WebAPIRequestRemoteInstanceList:
			remotesData := WebAPICollectRemoteInstances()

			c.WriteJSON(map[string]interface{}{
				"RequestType":     WebAPIAnswerRemoteInstanceList,
				"RemoteInstances": remotesData,
			})
			break
		}
	}
}

var WebAPIBroadcastSockets = make([]*websocket.Conn, 0)
var WebAPIBroadcastQueue = make(chan map[string]interface{}, 1024)

func WebAPIRegisterSocket(c *websocket.Conn) {
	// todo: mutex
	WebAPIBroadcastSockets = append(WebAPIBroadcastSockets, c)
	logger.Println("Info: Web Interface: Opening API Socket")
}

func WebAPIUnregisterSocket(conn *websocket.Conn) {
	// todo: mutex
	pos := -1
	for i, c := range WebAPIBroadcastSockets {
		if c == conn {
			pos = i
			break
		}
	}

	if pos != -1 {
		WebAPIBroadcastSockets = append(WebAPIBroadcastSockets[:pos], WebAPIBroadcastSockets[(pos+1):]...)
	}

	logger.Println("Info: Web Interface: Closed API Socket")
}

func WebAPICollectSensors() []map[string]interface{} {
	// todo: mutex

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

	return sensorData
}

func WebAPIBroadcastSensors() {
	WebAPIBroadcastQueue <- map[string]interface{}{
		"RequestType": WebAPIAnswerSensorList,
		"Sensors":     WebAPICollectSensors(),
	}
}

func WebAPICollectRemoteInstances() []map[string]interface{} {
	// todo: mutex

	remotesData := make([]map[string]interface{}, 0)
	for _, r := range local.RemoteInstances {
		remotesData = append(remotesData, map[string]interface{}{
			"UUID":        r.UUID,
			"DisplayName": r.DisplayName,
			"Sensors":     r.SensorUUIDs,
			"Connected":   r.connected,
		})
	}

	return remotesData
}

func WebAPIBroadcastRemoteInstances() {
	WebAPIBroadcastQueue <- map[string]interface{}{
		"RequestType":     WebAPIAnswerRemoteInstanceList,
		"RemoteInstances": WebAPICollectRemoteInstances(),
	}
}

func WebAPIBroadcast() {
	for {
		req := <-WebAPIBroadcastQueue

		// todo: mutex
		for _, c := range WebAPIBroadcastSockets {
			c.WriteJSON(req)
		}
	}
}
