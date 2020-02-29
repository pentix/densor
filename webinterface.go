package main

import (
	"html/template"
	"net/http"
)

func StartWebInterface() {
	http.HandleFunc("/", WebinterfaceRoot)
	http.ListenAndServeTLS("0.0.0.0:8334", local.DataDir+"cert.pem", local.DataDir+"key.pem", nil)
}

func WebinterfaceRoot(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("web/index.html")
	if err != nil {
		logger.Printf("Webinterface: Error: Could not load template:", err)
	}

	t.Execute(w, local)
}
