package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type JsonMessage struct {
	Payload string `json:"payload"`
	Topic   string `json:"topic"`
}

const pongWait = 60 * time.Second
const pingPeriod = (pongWait * 9) / 10

var upgrader = websocket.Upgrader{}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

var parseMessage = func(m mqtt.Message) *JsonMessage {
	return &JsonMessage{
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

	var subscribedMessages chan *JsonMessage = make(chan *JsonMessage)
	topic := path.Join(config.TopicPrefix, "#")
	client.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		subscribedMessages <- parseMessage(m)
	})

	go func() {
		var err error
		var jsonMessage *JsonMessage
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case jsonMessage = <-subscribedMessages:
				err = conn.WriteJSON(jsonMessage)
				switch err.(type) {
				case *websocket.CloseError:
					return
				default:
					fmt.Printf("Error writing to websocket: %v\n", err)
				}
			case <-ticker.C:
				err = conn.WriteMessage(websocket.PingMessage, nil)
				switch err.(type) {
				case *websocket.CloseError:
					return
				default:
					fmt.Printf("Error writing to websocket: %v\n", err)
				}
			}
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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
	default:
		fmt.Printf("Error reading from websocket: %v\n", err)
	}
}
