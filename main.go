package main

import (
	"fmt"
	"net/http"
)

var config ApiConfig

func checkOrigin(r *http.Request) bool {
	var forwardedHost = r.Header.Get("X-Forwarded-Host")
	var forwardedProto = r.Header.Get("X-Forwarded-Proto")
	switch forwardedProto {
	case "ws":
		forwardedProto = "http"
	case "wss":
		forwardedProto = "https"
	}

	if forwardedHost != "" && forwardedProto != "" {
		var forwardedOrigin = forwardedProto + "://" + forwardedHost
		if r.Header.Get("Origin") == forwardedOrigin {
			return true
		}
	}

	for _, corsOrigin := range config.CorsOrigins {
		if r.Header.Get("Origin") == corsOrigin {
			return true
		}
	}

	return false
}

func main() {
	getConfig(&config)
	upgrader.CheckOrigin = checkOrigin

	http.HandleFunc("/ws/mqtt", websocketHandler)
	fmt.Println("Listening on:", config.BindAddress)
	http.ListenAndServe(config.BindAddress, nil)
}
