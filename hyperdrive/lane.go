package hyperdrive

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	LaneTopic       = "RemoteControl/U/E/vehicles/%s/lane"
	CancelLaneTopic = "RemoteControl/U/E/vehicles/%s/cancelLane"
)

// LanePayload correspond à la structure LaneIntentStatus
type LanePayload struct {
	Velocity         float32 `json:"velocity"`         // {0...1000}
	Acceleration     float32 `json:"acceleration"`     // {0...2000}
	Offset           float32 `json:"offset"`           // {-100...100}
	OffsetFromCenter float32 `json:"offsetFromCenter"` // {-100...100}
}

// CancelLanePayload correspond à la structure CancelLaneIntentStatus
type CancelLanePayload struct {
	Value bool `json:"value"` // {true|false}
}

// lane envoie une commande de changement de piste.
// Le changement est implicite dans les valeurs de OffsetFromCenter ou Offset.
func lane(client mqtt.Client, velocity float32, acceleration float32, offsetFromCenter float32, offset float32, target string) error {
	payload, err := json.Marshal(LanePayload{
		Velocity:         velocity,
		Acceleration:     acceleration,
		OffsetFromCenter: offsetFromCenter,
		Offset:           offset,
	})
	if err != nil {
		return err
	}

	laneTopic := fmt.Sprintf(LaneTopic, target)
	log.Println("[Lane] Sending", string(payload), "on", laneTopic)

	if token := client.Publish(laneTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// cancelLane envoie un message pour annuler le changement de piste en cours.
func cancelLane(client mqtt.Client, target string) error {
	payload, err := json.Marshal(CancelLanePayload{
		Value: true, // Pour annuler, on envoie généralement true
	})
	if err != nil {
		return err
	}

	cancelLaneTopic := fmt.Sprintf(CancelLaneTopic, target)
	log.Println("[Lane] Sending cancel payload on", cancelLaneTopic)

	if token := client.Publish(cancelLaneTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
