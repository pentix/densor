package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func StartWebInterface() {
	http.HandleFunc("/", WebUIRoot)
	http.HandleFunc("/login", WebUILogin)
	http.HandleFunc("/logout", WebUILogout)
	http.HandleFunc("/api", WebAPI)

	go WebAPIBroadcast()
	http.ListenAndServeTLS("0.0.0.0:8334", local.config.GetString("WebTLSCert"), local.config.GetString("WebTLSKey"), nil)
}

func WebUIRoot(w http.ResponseWriter, req *http.Request) {
	if !WebUILoggedIn(req) {
		WebUILogin(w, req)
		return
	}

	t, err := template.ParseFiles("web/index.html") // todo: parse only once
	if err != nil {
		logger.Printf("Error: Web Interface: Could not load template:", err, err)
		return
	}

	t.Execute(w, local)
}

func WebUILoggedIn(req *http.Request) bool {
	c, err := req.Cookie("sessionID")
	if err != nil {
		return false
	}

	return WebUICheckSession(c.Value)
}

func WebUILogin(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		logger.Println("Error: Web Interface: Error parsing HTTP data :", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := req.Form.Get("username")
	password := req.Form.Get("password")
	if username == "" || password == "" {
		t, err := template.ParseFiles("web/login.html") // todo: parse only once
		if err != nil {
			logger.Printf("Error: Web Interface: Could not load template:", err, err)
			return
		}

		t.Execute(w, nil)
		return
	}

	if username == "test" && password == "demo" {
		http.SetCookie(w, &http.Cookie{
			Name:     "sessionID",
			Value:    "asdfasdf",
			Path:     "",
			Expires:  time.Now().Add(2 * time.Hour),
			MaxAge:   2 * 60 * 60,
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return

	} else {

		t, err := template.ParseFiles("web/login.html") // todo: parse only once
		if err != nil {
			logger.Printf("Error: Web Interface: Could not load template:", err, err)
			return
		}

		t.Execute(w, WebUIMessage{
			Message: "Check your credentials. Wrong user/password",
			isError: true,
		})
		return
	}
}

func WebUILogout(w http.ResponseWriter, req *http.Request) {
	// todo: unregister session and remove cookie from header
	http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
	return
}

type WebUIMessage struct {
	Message string
	isError bool
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
