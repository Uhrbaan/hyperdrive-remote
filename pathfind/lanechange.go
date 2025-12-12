package main

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/*

{
	ID: "anything not null"
	lane_change: "left" | "right" | "" (null),
	forward: true | false
}

*/

const (
	LaneTopic  = "Anki/Vehicles/%s/S/intended/lane"
	SpeedTopic = "Anki/Vehicles/%s/S/intended/speed"
	LaneChangeTopic = "PlaceHolder/S/LaneChangeTopic"
	LaneDirValue = 68
	AccelerationValue = 300
	VelocityValue = 300
	NegVelocityValue = -100
)

type LanePayload struct {
	Velocity         float32 `json:"velocity"`         // {0...1000}
	Acceleration     float32 `json:"acceleration"`     // {0...2000}
	Offset           float32 `json:"offset"`           // {-100...100}
	OffsetFromCenter float32 `json:"offsetFromCenter"` // {-100...100}
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

func lane_change(client mqtt.Client, velocity float32, acceleration float32, offsetFromCenter float32, offset float32, target string) error {
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
	log.Println("[Speed] Sending", string(payload), "on", SpeedTopic)

	if token := client.Publish(speedTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return err
	}

	return nil
}

func laneChangeHandler(client mqtt.Client, msg mqtt.Message) {
	var lcMsg LaneChangeMessage
	err := json.Unmarshal(msg.Payload(), &lcMsg)
	if err != nil {
		log.Println("[LaneChange] Error decoding message:", err)
		return
	}

	// Access the fields
	log.Println("[LaneChange] Received message from ID:", lcMsg.ID)
	log.Println("[LaneChange] Lane change:", lcMsg.LaneChange)
	log.Println("[LaneChange] Forward:", lcMsg.Forward)

	if lcMsg.Forward {
    	offsetFromCenter := 0

    	if lcMsg.LaneChange == "left" {
     	   offsetFromCenter = -LaneDirValue
    	} else if lcMsg.LaneChange == "right" {
     	   offsetFromCenter = LaneDirValue
    	}

    err := lane_change(client, VelocityValue, AccelerationValue, float32(offsetFromCenter), 0, lcMsg.ID)
    if err != nil {
        log.Println("[Lane] Error sending lane command:", err)
    	}
	}

	else {
		    err := speed(client, NegVelocityValue, AccelerationValue, lcMsg.ID)
    if err != nil {
        log.Println("[Speed] Error sending speed command:", err
		}
	}

func subscribeLaneChange(client mqtt.Client) {
	if token := client.Subscribe(LaneChangeTopic, 1, laneChangeHandler); token.Wait() && token.Error() != nil {
		log.Fatal("[LaneChange] Subscribe error:", token.Error())
	}
	log.Println("[LaneChange] Subscribed to topic:", LaneChangeTopic)
}















//-----------------------------------------------------------------------------------------------------------

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("myClientID")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// Subscribe to lane change messages
	subscribeLaneChange(client)

	// Keep the program running
	select {}
}
