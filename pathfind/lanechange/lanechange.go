package lanechange

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	LaneTopic         = "Anki/Vehicles/%s/S/intended/lane"
	SpeedTopic        = "Anki/Vehicles/%s/S/intended/speed"
	LaneChangeTopic   = "PlaceHolder/S/LaneChangeTopic"
	LaneDirValue      = 68
	AccelerationValue = 300
	VelocityValue     = 300
	NegVelocityValue  = -100
)

type LanePayload struct {
	Velocity         float32 `json:"velocity"`
	Acceleration     float32 `json:"acceleration"`
	Offset           float32 `json:"offset"`
	OffsetFromCenter float32 `json:"offsetFromCenter"`
}

type SpeedPayload struct {
	Velocity     float32 `json:"velocity"`
	Acceleration float32 `json:"acceleration"`
}

type LaneChangeMessage struct {
	ID         string `json:"ID"`
	LaneChange string `json:"lane_change"`
	Forward    bool   `json:"forward"`
}

func lane_change(
	client mqtt.Client,
	velocity float32,
	acceleration float32,
	offsetFromCenter float32,
	offset float32,
	target string,
) error {

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
		return token.Error()
	}
	return nil
}

func laneChangeHandler(client mqtt.Client, msg mqtt.Message) {
	var lcMsg LaneChangeMessage
	if err := json.Unmarshal(msg.Payload(), &lcMsg); err != nil {
		log.Println("[LaneChange] Error decoding message:", err)
		return
	}

	log.Println("[LaneChange] Received message from ID:", lcMsg.ID)
	log.Println("[LaneChange] Lane change:", lcMsg.LaneChange)
	log.Println("[LaneChange] Forward:", lcMsg.Forward)

	// Forward lane change
	if lcMsg.Forward {
		offsetFromCenter := 0

		if lcMsg.LaneChange == "left" {
			offsetFromCenter = -LaneDirValue
		} else if lcMsg.LaneChange == "right" {
			offsetFromCenter = LaneDirValue
		}

		if err := lane_change(
			client,
			VelocityValue,
			AccelerationValue,
			float32(offsetFromCenter),
			0,
			lcMsg.ID,
		); err != nil {
			log.Println("[Lane] Error sending lane command:", err)
		}
	}

	// Reverse / stop
	if !lcMsg.Forward {
		if err := speed(client, NegVelocityValue, AccelerationValue, lcMsg.ID); err != nil {
			log.Println("[Speed] Error sending speed command:", err)
		}
	}
}

func subscribeLaneChange(client mqtt.Client) {
	if token := client.Subscribe(LaneChangeTopic, 1, laneChangeHandler); token.Wait() && token.Error() != nil {
		log.Fatal("[LaneChange] Subscribe error:", token.Error())
	}
	log.Println("[LaneChange] Subscribed to topic:", LaneChangeTopic)
}

func main() {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("myClientID")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	subscribeLaneChange(client)

	// Keep running
	select {}
}
