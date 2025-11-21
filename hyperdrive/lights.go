package hyperdrive

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	LightsTopic = "RemoteControl/U/E/vehicles/%s/lights"
)

type LightEffect struct {
	Effect    string `json:"effect"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Frequency int    `json:"frequency"`
}

type LightPayload struct {
	FrontGreen  LightEffect `json:"frontGreen"`
	FrontRed    LightEffect `json:"frontRed"`
	Tail        LightEffect `json:"tail"`
	EngineRed   LightEffect `json:"engineRed"`
	EngineGreen LightEffect `json:"engineGreen"`
	EngineBlue  LightEffect `json:"engineBlue"`
}

func lights(client mqtt.Client, params LightPayload, target string) error {
	payload, err := json.Marshal(params)
	if err != nil {
		return err
	}

	lightsTopic := fmt.Sprintf(LightsTopic, target)
	log.Println("[Speed] Sending", string(payload), "on", lightsTopic)

	if token := client.Publish(lightsTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}
