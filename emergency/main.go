package main

import (
	"encoding/json"
	"flag"
	"hyperdrive/remote/hyperdrive"
	"log"
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
	stopTopic = "Emergency/U/E/stop"
)

// Configuration des variables pour le broker MQTT, l'ID client et le QoS.
var (
	brokerHost   = flag.String("broker", "10.42.0.1:1883", "MQTT broker URL")
	clientIDFlag = flag.String("id", "hjkàélkgjhjvbkljgfjxdgchjkl", "Client ID for Emergency (default: random UUID)")
	qosFlag      = flag.Int("qos", 1, "MQTT QoS")
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
	client      mqtt.Client // MQTT client
	id          string      // Client ID
	qos         byte        // QoS level
	stop        bool        // Indique si le mode d'arrêt d'urgence est actif
	subTopicIDs []string    // Liste des topics abonnés (non utilisé dans cette version)
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

// publishIntentToVehicles permet de diffuser un Intent à tous les véhicules (Anki/Vehicles/U/I/<callerID>)

// handleStopMessage : gestionnaire de messages pour Emergency/U/E/stop
func (e *Emergency) publishStopMessage(client mqtt.Client) {
	// Si le mode d'arrêt est activé, publier immédiatement un intent speed=0 à tous les véhicules
	if !e.stop {
		return
	}

	intent := Intent{
		Type: "speed",
		Payload: map[string]interface{}{
			"velocity":     0,
			"acceleration": 0,
		},
	}
	log.Println("Emergency: publishing immediate speed=0 to all vehicles")

	data, err := json.Marshal(intent)
	if err != nil {
		return
	}

	client.Publish(stopTopic, 1, false, data)
}

// ------------------------------------------------------------------------------------------
// Voir à partir d'ici pour le relais des messages RemoteControl
// ------------------------------------------------------------------------------------------

// mapRemoteTopicToMediate crée le topic mediate pour un topic RemoteControl donné
func mapRemoteTopicToMediate(remoteTopic string) string {
	return "Emergency/U/E/mediate-De-Kevin-et-Leonard/" + remoteTopic
}

// handleRemoteVehicleMessage sert de gestionnaire pour les messages RemoteControl véhicules
// Il relaie les messages sauf si stopActive. Il mappe simplement les topics RemoteControl vers Emergency/U/E/mediate/...
// func (e *Emergency) handleRemoteVehicleMessage(client mqtt.Client, msg mqtt.Message) {
// 	log.Println("Got message from", msg.Topic(), "mirroring to", mapRemoteTopicToMediate(msg.Topic()))
// 	if e.stop == true {
// 		log.Printf("Emergency: STOP active, ignoring remote message on %s", msg.Topic())
// 		return
// 	} else {
// 		log.Println("Should be disabled.")
// 	}

// 	mediateTopic := mapRemoteTopicToMediate(msg.Topic())
// 	token := client.Publish(mediateTopic, 1, false, msg.Payload())
// 	token.Wait()
// 	if token.Error() != nil {
// 		log.Fatal("Something terrible happened while mirroring remote: failed to publish:", token.Error())
// 	}

// 	log.Printf("Emergency: forwarded %s -> %s", msg.Topic(), mediateTopic)
// }

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

	// Souscrire aux événements des véhicules RemoteControl
	if tok := client.Subscribe("RemoteControl/#", 1, func(client mqtt.Client, msg mqtt.Message) {
		log.Println("Got message from", msg.Topic(), "mirroring to", mapRemoteTopicToMediate(msg.Topic()))
		if em.stop == true {
			log.Printf("Emergency: STOP active, ignoring remote message on %s", msg.Topic())
			return
		}

		mediateTopic := mapRemoteTopicToMediate(msg.Topic())
		if token := client.Publish(mediateTopic, 1, false, msg.Payload()); token.Error() != nil {
			log.Fatal("Something terrible happened while mirroring remote: failed to publish:", token.Error())
		}

		log.Printf("Emergency: forwarded %s -> %s", msg.Topic(), mediateTopic)

	}); tok.Wait() && tok.Error() != nil {
		log.Fatalf("Subscribe to remote vehicles failed: %v", tok.Error())
	}

	// APPLICATION

	// Create app
	isStopped := binding.NewBool()

	w := app.New().NewWindow("title")
	w.Resize(fyne.NewSize(350, 180))
	isStopped.Set(false)

	hostIntentTopicEntry := widget.NewEntry()
	hostIntentTopicEntry.SetText("Anki/Hosts/U/I")
	vehicleIntentTopicFormatEntry := widget.NewEntry()
	vehicleIntentTopicFormatEntry.SetText("Anki/Vehicles/U/%s/I")
	hostDiscoverVehicleTopicEntry := widget.NewEntry()
	hostDiscoverVehicleTopicEntry.SetText("Anki/Hosts/U/hyperdrive/E/vehicle/discovered/#")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Topic for the Anki Host intent:",
				Widget: hostIntentTopicEntry,
			},
			{
				Text:   "Topic format (where %s is the car id) of the car intents:",
				Widget: vehicleIntentTopicFormatEntry,
			},
			{
				Text:   "Topic where to discover vehicles",
				Widget: hostDiscoverVehicleTopicEntry,
			},
		},
		OnSubmit: func() {
			vehicleMap, _ := hyperdrive.Discover(client, hostDiscoverVehicleTopicEntry.Text)
			log.Println(vehicleMap)

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
