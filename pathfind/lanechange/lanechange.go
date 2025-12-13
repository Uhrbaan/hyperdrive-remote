package lanechange

import (
	"encoding/json"
	"fmt"
	"hyperdrive/remote/hyperdrive"
	"hyperdrive/remote/pathfind/util"
	"log"
	"slices"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	laneTopic         = util.RootTopic + "/lane"
	speedTopic        = util.RootTopic + "/speed"
	connectTopic      = util.RootTopic + "/connect"
	InstructionTopic  = util.RootTopic + "/instruction"
	AnkiVehicleIntent = "Anki/Vehicles/U/%s/I"

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
	LaneChange string `json:"lane_change"`
	Forward    bool   `json:"forward"`
}

func lane_change(
	client mqtt.Client,
	velocity float32,
	acceleration float32,
	offsetFromCenter float32,
	offset float32,
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

	log.Println("[Lane] Sending", string(payload), "on", laneTopic)

	if token := client.Publish(laneTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func speed(client mqtt.Client, velocity float32, acceleration float32) error {
	payload, err := json.Marshal(SpeedPayload{
		Velocity:     velocity,
		Acceleration: acceleration,
	})
	if err != nil {
		return err
	}

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

	log.Println("[LaneChange] Lane change:", lcMsg.LaneChange)
	log.Println("[LaneChange] Forward:", lcMsg.Forward)

	// Forward lane change
	if lcMsg.Forward {
		offsetFromCenter := 0

		switch lcMsg.LaneChange {
		case "left":
			offsetFromCenter = -LaneDirValue
		case "right":
			offsetFromCenter = LaneDirValue
		}

		if err := lane_change(
			client,
			VelocityValue,
			AccelerationValue,
			float32(offsetFromCenter),
			0,
		); err != nil {
			log.Println("[Lane] Error sending lane command:", err)
		}
	}

	// Reverse / stop
	if !lcMsg.Forward {
		if err := speed(client, NegVelocityValue, AccelerationValue); err != nil {
			log.Println("[Speed] Error sending speed command:", err)
		}
	}
}

func subscribeLaneChange(client mqtt.Client) {
	if token := client.Subscribe(InstructionTopic, 1, laneChangeHandler); token.Wait() && token.Error() != nil {
		log.Fatal("[LaneChange] Subscribe error:", token.Error())
	}
	log.Println("[LaneChange] Subscribed to topic:", InstructionTopic)
}

func connectEverything(client mqtt.Client, id string) {
	stepCh := make(chan struct{})
	baseAnkiSub := "Anki/Vehicles/U/%s/S/DIT/%s"
	client.Subscribe(fmt.Sprintf(baseAnkiSub, id, "connectSubscription"), 1, func(c mqtt.Client, m mqtt.Message) {
		var data map[string]any
		json.Unmarshal(m.Payload(), &data)
		if data["value"] != nil && slices.Contains(data["value"].([]string), connectTopic) {
			stepCh <- struct{}{}
		}
	})
	hyperdrive.SyncSubscription(client, "connectSubscription", fmt.Sprintf(AnkiVehicleIntent, id), connectTopic, true)
	<-stepCh // wait for the subscription to go through

	client.Subscribe(fmt.Sprintf(baseAnkiSub, id, "speedSubscription"), 1, func(c mqtt.Client, m mqtt.Message) {
		var data map[string]any
		json.Unmarshal(m.Payload(), &data)
		if slices.Contains(data["value"].([]string), speedTopic) {
			stepCh <- struct{}{}
		}
	})
	hyperdrive.SyncSubscription(client, "speedSubscription", fmt.Sprintf(AnkiVehicleIntent, id), speedTopic, true)
	<-stepCh

	client.Subscribe(fmt.Sprintf(baseAnkiSub, id, "laneSubscription"), 1, func(c mqtt.Client, m mqtt.Message) {
		var data map[string]any
		json.Unmarshal(m.Payload(), &data)
		if slices.Contains(data["value"].([]string), laneTopic) {
			stepCh <- struct{}{}
		}
	})
	hyperdrive.SyncSubscription(client, "laneSubscription", fmt.Sprintf(AnkiVehicleIntent, id), connectTopic, true)
	<-stepCh

	log.Println("")
}

func InstructionProcess(client mqtt.Client) {
	// 1. get vehicle ID from UI
	vehicleID := util.WaitForVehicleID(client)

	// 2. Publish necessary instructions for DIT to work (connect, speed, lane)
	connectEverything(client, vehicleID)

	// 3. Connect to the vehicle and publish initial speed instruction
	data, _ := json.Marshal(hyperdrive.ConnectPayload{Value: true})
	client.Publish(connectTopic, 1, false, data)
	speed(client, VelocityValue, AccelerationValue)

	subscribeLaneChange(client)

	select {} // block indefinetely
}

func main() {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("myClientID")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// SubscribeLaneChange(client, LaneChangeTopic)

	// Keep running
	select {}
}
