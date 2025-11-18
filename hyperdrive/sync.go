package hyperdrive

import (
	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/*
NOTE: this file is technically not part of the project.
It simply contains a few utilities used to do all the subscriptions between the Remote and the Vehicles.
*/

type Subscription struct {
	Topic     string `json:"topic"`
	Subscribe bool   `json:"subscribe"`
}

// client: a mosquitto client
// subscriptionType: connect|lights|...
// subscriptionTargetTopic: topic where to publish the subscription
// topic: the topic the target should subscribe to
// subscribe: boolean whether to enable or disable the subscription
func SyncSubscription(client mqtt.Client, subscriptionType string, subscriptionTargetTopic string, topic string, subscribe bool) error {
	data, err := json.Marshal(Intent{
		Type: subscriptionType,
		Payload: Subscription{
			Topic:     topic,
			Subscribe: subscribe,
		},
	})
	/*
		{

		}
	*/

	if err != nil {
		return err
	}

	if token := client.Publish(subscriptionTargetTopic, 1, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
