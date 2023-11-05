package main

import (
	"encoding/json"
	"os"
)

type ApiConfig struct {
	BindAddress string           `json:"bindAddress"`
	CorsOrigins []string         `json:"corsOrigins"`
	MqttServer  MqttServerConfig `json:"mqttServer"`
	TopicPrefix string           `json:"topicPrefix"`
}

type MqttServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func getConfig(config *ApiConfig) {
	var buffer = make([]byte, 1024)
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	length, err := file.Read(buffer)
	if err != nil {
		panic(err)
	}

	buffer = buffer[0:length]
	err = json.Unmarshal(buffer, &config)
	if err != nil {
		panic(err)
	}
}
