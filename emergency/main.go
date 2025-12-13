package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hyperdrive/remote/hyperdrive"
	"log"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	stopTopic        = "Emergency/U/E/stop"
	mediateRootTopic = "Emergency/U/E/mediate/"
)

// Configuration des variables pour le broker MQTT, l'ID client et le QoS.
var (
	brokerHost                = flag.String("broker", "10.42.0.1:1883", "MQTT broker URL")
	clientIDFlag              = flag.String("id", "kevin-leo-emergency-control", "Client ID for Emergency (default: random UUID)")
	qosFlag                   = flag.Int("qos", 1, "MQTT QoS")
	vehicleSubscriptionFormat string
	remoteInstructionsFormat  string
)

// Definition de la structure Intent pour les messages publiés aux véhicules et hôtes.
type Intent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// stopMessagePayload est une structure simple que nous attendons sur Emergency/U/E/stop.
type stopMessagePayload struct {
	Value bool `json:"value"`
}

// Emergency gère l'état d'arrêt d'urgence et relaie les messages MQTT.
type Emergency struct {
	client      mqtt.Client         // MQTT client
	id          string              // Client ID
	qos         byte                // QoS level
	stop        bool                // Indique si le mode d'arrêt d'urgence est actif
	vehicleList map[string][]string // Liste des topics abonnés (non utilisé dans cette version)
}

// NewEmergency crée une nouvelle instance d'Emergency.
func NewEmergency(client mqtt.Client, id string, qos byte) *Emergency {
	return &Emergency{ // Initialisation de la structure Emergency
		client: client, // Assignation du client MQTT
		id:     id,     // Assignation de l'ID client
		qos:    qos,    // Assignation du niveau de QoS
		stop:   false,
	}
}

// handleStopMessage : gestionnaire de messages pour Emergency/U/E/stop
func (e *Emergency) publishStopMessage(client mqtt.Client) {
	// Si le mode d'arrêt est activé, publier immédiatement un intent speed=0 à tous les véhicules
	if !e.stop {
		return
	}

	intent := hyperdrive.SpeedPayload{
		Velocity:     0,
		Acceleration: 1000,
	}
	log.Println("Emergency: publishing immediate speed=0 to all vehicles")

	data, err := json.Marshal(intent)
	if err != nil {
		return
	}

	if token := client.Publish(stopTopic, 1, false, data); token.Wait() && token.Error() != nil {
		log.Println("[Emergency] Got error while sending stop:", token.Error())
	}
}

// mapRemoteTopicToMediate crée le topic mediate pour un topic RemoteControl donné
func mapRemoteTopicToMediate(remoteTopic string) string {
	return mediateRootTopic + remoteTopic
}

// Fonction principale qui permet de configurer le client MQTT, de s'abonner aux topics nécessaires et de gérer la boucle principale.
func main() {
	flag.Parse()
	log.Println("Got", *brokerHost, "as the broker url")
	log.Println("Got", *clientIDFlag, "as id")
	log.Println("Got", *qosFlag, "as quality of service")

	id := *clientIDFlag
	if id == "" {
		id = "Emergency-" + uuid.NewString()
	}
	qos := byte(*qosFlag)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*brokerHost)
	opts.SetClientID(id)
	// Auto-reconnect
	opts.AutoReconnect = true
	opts.ConnectTimeout = 5 * time.Second
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	}
	opts.OnConnect = func(c mqtt.Client) {
		log.Printf("MQTT connected (client id=%s)", id)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Could not connect to broker: %v", token.Error())
	}
	em := NewEmergency(client, id, qos)

	// Create app
	isStopped := binding.NewBool()

	w := app.New().NewWindow("Emergency")
	w.Resize(fyne.NewSize(350, 180))
	isStopped.Set(false)

	vehicleIntentTopicFormatEntry := widget.NewEntry()
	vehicleIntentTopicFormatEntry.SetText("Anki/Vehicles/U/%s/I")
	remoteVehicleInstructionsTopicEntry := widget.NewEntry()
	remoteVehicleInstructionsTopicEntry.SetText("RemoteControl/U/E/vehicles/%8s/%s")
	remoteRootTopicEntry := widget.NewEntry()
	remoteRootTopicEntry.SetText("RemoteControl/#")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Topic format (where %s is the car id) of the car intents:",
				Widget: vehicleIntentTopicFormatEntry,
			},
			{
				Text:   "First %s for the vehicle, second for the subscription type",
				Widget: remoteVehicleInstructionsTopicEntry,
			},
			{
				Text:   "RemoteControl root topic",
				Widget: remoteRootTopicEntry,
			},
		},
		OnSubmit: func() {
			vehicleSubscriptionFormat = vehicleIntentTopicFormatEntry.Text
			remoteInstructionsFormat = remoteVehicleInstructionsTopicEntry.Text

			// Souscrire aux événements des véhicules RemoteControl
			if tok := client.Subscribe(remoteRootTopicEntry.Text, 1, func(client mqtt.Client, msg mqtt.Message) {
				log.Println("Got message from", msg.Topic(), "mirroring to", mapRemoteTopicToMediate(msg.Topic()))
				if em.stop == true {
					log.Printf("Emergency: STOP active, ignoring remote message on %s", msg.Topic())
					return
				}

				var vehicleID string
				var payloadType string
				n, err := fmt.Sscanf(msg.Topic(), remoteInstructionsFormat, &vehicleID, &payloadType)
				log.Println("Got vehicle", vehicleID, "for the payload type", payloadType)
				if n == 2 && err != nil {
					log.Fatalf("The format provided: %s is not correct: %v", remoteInstructionsFormat, err)
				}

				mediateTopic := mapRemoteTopicToMediate(msg.Topic())

				// Initialize the map if not already.
				if em.vehicleList == nil {
					em.vehicleList = map[string][]string{}
				}

				subscriptionType, exists := em.vehicleList[vehicleID]

				// If the car does not exist, or if the passed type is not in the list
				if !exists || !slices.Contains(subscriptionType, payloadType) {
					// If the car does not exist, add it to the list with the payload type.
					if !exists {
						em.vehicleList[vehicleID] = []string{payloadType}

						// Give it the stop topic directly upon creation
						err = hyperdrive.SyncSubscription(client, "speedSubscription", fmt.Sprintf(vehicleSubscriptionFormat, vehicleID), stopTopic, true)
						if err != nil {
							log.Println("Failed to subscribe the vehicle to the emergency stop:", err)
						}
						log.Println("Successfully sent Stop subscription", stopTopic, "to", fmt.Sprintf(vehicleSubscriptionFormat, vehicleID))
						time.Sleep(1000 * time.Millisecond) // enusure subscription gets registerd before sending the next
					}

					// if the car exists, but the payload type is still unknown, then add the payload type.
					if !slices.Contains(subscriptionType, payloadType) {
						em.vehicleList[vehicleID] = append(em.vehicleList[vehicleID], payloadType)
					}

					// Subscribe to the suscription type that was sent
					err := hyperdrive.SyncSubscription(client, payloadType+"Subscription", fmt.Sprintf(vehicleSubscriptionFormat, vehicleID), mediateTopic, true)
					if err != nil {
						log.Println("Failed to subscribe the vehicle the emergency remote:", err)
					}
					log.Println("Successfully sent subscription of", mediateTopic, "to", fmt.Sprintf(vehicleSubscriptionFormat, vehicleID))
					time.Sleep(1000 * time.Millisecond) // enusure subscription gets registerd before sending the next
				}

				if token := client.Publish(mediateTopic, 1, false, msg.Payload()); token.Error() != nil {
					log.Fatal("Something terrible happened while mirroring remote: failed to publish:", token.Error())
				}

				log.Printf("Emergency: forwarded %s -> %s", msg.Topic(), mediateTopic)

			}); tok.Wait() && tok.Error() != nil {
				log.Fatalf("Subscribe to remote vehicles failed: %v", tok.Error())
			}

			statusLabel := widget.NewLabelWithData(binding.BoolToString(isStopped))

			stopButton := widget.NewButton("Stop", func() {
				// Log the action to the console (optional)
				println("Stop requested: setting state to true")
				em.stop = true
				em.publishStopMessage(client)
				isStopped.Set(true)
			})

			continueButton := widget.NewButton("Continue", func() {
				// Log the action to the console (optional)
				println("Continue requested: setting state to false")
				em.stop = false
				em.publishStopMessage(client)
				isStopped.Set(false)
			})

			buttonContainer := container.NewGridWithColumns(2, stopButton, continueButton)

			content := container.NewVBox(
				statusLabel,
				layout.NewSpacer(), // Pushes the label up a bit
				buttonContainer,
			)

			// replace the form by the cars
			w.SetContent(content)
		},
	}

	w.SetContent(form)

	w.ShowAndRun()
}
