package hyperdrive

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ConnectPayload struct {
	Value bool `json:"value"` // {true|false} # Default: false
}

const (
	ConnectTopic = "RemoteControl" + UserSuffix + "/U/E/vehicles/%s/connect"
)

// Send a connect payload to list of targets
func connect(client mqtt.Client, value bool, target string) error {
	payload, err := json.Marshal(ConnectPayload{
		Value: value,
	})

	if err != nil {
		return err
	}

	connectTopic := fmt.Sprintf(ConnectTopic, target)
	log.Println("[Connect] Sending", string(payload), "on", connectTopic)

	if token := client.Publish(connectTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
