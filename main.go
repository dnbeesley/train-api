package main

import (
	"fmt"
	"net/http"
)

var config ApiConfig

func main() {
	getConfig(&config)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	http.HandleFunc("/ws/mqtt", websocketHandler)
	fmt.Println("Listening on:", config.BindAddress)
	http.ListenAndServe(config.BindAddress, nil)
}
