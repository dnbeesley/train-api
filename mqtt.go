package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type JsonMessage struct {
	Payload string `json:"payload"`
	Topic   string `json:"topic"`
}

var upgrader = websocket.Upgrader{}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

var getMessageHandler = func(conn *websocket.Conn) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		conn.WriteJSON(parseMessage(m))
	}
}

var parseMessage = func(m mqtt.Message) JsonMessage {
	return JsonMessage{
		Payload: string(m.Payload()),
		Topic:   m.Topic(),
	}
}

var websocketHandler = func(w http.ResponseWriter, r *http.Request) {
	var id = uuid.New()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrade failed: %v\n", err)
		w.WriteHeader(500)
		return
	}

	opts := mqtt.NewClientOptions()
	server := fmt.Sprintf("tcp://%s:%d", config.MqttServer.Host, config.MqttServer.Port)
	fmt.Println("Connecting to:", server)
	opts.AddBroker(server)
	opts.SetClientID(path.Join(config.MqttServer.Username, id.String()))
	opts.SetUsername(config.MqttServer.Username)
	opts.SetPassword(config.MqttServer.Password)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Error connecting to MQTT: %v\n", token.Error())
		conn.Close()
		return
	}

	topic := path.Join(config.TopicPrefix, "#")
	client.Subscribe(topic, 1, getMessageHandler(conn))

	var payload []byte
	for err == nil {
		_, payload, err = conn.ReadMessage()
		var jsonMessage JsonMessage
		json.Unmarshal(payload, &jsonMessage)
		if !strings.HasPrefix(jsonMessage.Topic, config.TopicPrefix) {
			continue
		}

		if token := client.Publish(jsonMessage.Topic, 0, true, []byte(jsonMessage.Payload)); token.Wait() && token.Error() != nil {
			fmt.Printf("Error connecting to MQTT: %v\n", token.Error())
			conn.Close()
			return
		}
	}

	switch err.(type) {
	case *websocket.CloseError:
		client.Disconnect(1000)
		break
	default:
		fmt.Printf("Error reading from websocket: %v\n", err)
	}
}
