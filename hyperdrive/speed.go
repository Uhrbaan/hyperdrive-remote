package hyperdrive

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	SpeedTopic = "RemoteControl/U/E/vehicles/%s/speed"
)

type SpeedPayload struct {
	Velocity     float32 `json:"velocity"`     // {-100...1000} # Default: 0
	Acceleration float32 `json:"acceleration"` // {0...2000} # Default: 0
}

func speed(client mqtt.Client, velocity float32, acceleration float32, target string) error {
	payload, err := json.Marshal(SpeedPayload{
		Velocity:     velocity,
		Acceleration: acceleration,
	})
	if err != nil {
		return err
	}

	speedTopic := fmt.Sprintf(SpeedTopic, target)
	log.Println("[Speed] Sending", string(payload), "on", speedTopic)

	if token := client.Publish(speedTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}
