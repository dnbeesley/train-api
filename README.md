# Train API

This program is designed to provide Web API to to publish and receive MQTT messages.

## Configuration

Configuration file, config.json, needs to be in the applications context path. The file needs to define the following:

```JSON
{
  "bindAddress": "localhost:8080",
  "mqttServer": {
    "username": "username",
    "password": "password",
    "host": "192.168.1.2",
    "port": 1883
  },
  "topicPrefix": "i2c-relay"
}
```

This example file will allow HTTP client to publish or subscribe to allow topics that match the pattern i2c-relay/# subject to the access controls of the user specified by the MQTT settings.
