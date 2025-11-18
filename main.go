package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// Configuration des variables pour le broker MQTT, l'ID client et le QoS.
var (
	brokerHost   = flag.String("broker", "10.42.0.1:1883", "MQTT broker URL")
	clientIDFlag = flag.String("id", "", "Client ID for Emergency (default: random UUID)")
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
	client     mqtt.Client  // MQTT client
	id         string       // Client ID
	qos        byte         // QoS level
	stopActive bool         // Indique si le mode d'arrêt d'urgence est actif
	stopMu     sync.RWMutex // Mutex pour protéger l'accès à stopActive
	// subTopicIDs []string     // Liste des topics abonnés (non utilisé dans cette version)
}

// NewEmergency crée une nouvelle instance d'Emergency.
func NewEmergency(client mqtt.Client, id string, qos byte) *Emergency {
	return &Emergency{ // Initialisation de la structure Emergency
		client: client, // Assignation du client MQTT
		id:     id,     // Assignation de l'ID client
		qos:    qos,    // Assignation du niveau de QoS
	}
}

// setStop définit l'état d'arrêt d'urgence.
func (e *Emergency) setStop(active bool) {
	e.stopMu.Lock() // Verrouillage en écriture
	defer e.stopMu.Unlock()
	e.stopActive = active // Mise à jour de l'état d'arrêt d'urgence
}

// isStop retourne true si le mode d'arrêt d'urgence est actif.
func (e *Emergency) isStop() bool {
	e.stopMu.RLock() // Verrouillage en lecture
	defer e.stopMu.RUnlock()
	return e.stopActive // Retourne l'état d'arrêt d'urgence
}

// publishIntentToVehicles permet de diffuser un Intent à tous les véhicules (Anki/Vehicles/U/I/<callerID>)
func (e *Emergency) publishIntentToVehicles(intent Intent) {
	topic := fmt.Sprintf("Anki/Vehicles/U/I/%s", e.id) // Construction du topic
	bs, _ := json.Marshal(intent)                      // Sérialisation de l'intent en JSON
	token := e.client.Publish(topic, e.qos, false, bs) // Publication du message
	token.WaitTimeout(3 * time.Second)                 // Attente de la confirmation de publication
	if token.Error() != nil {                          // Gestion des erreurs
		log.Printf("publishIntentToVehicles error: %v", token.Error())
	}
}

// publishIntentToVehicleTarget publie un Intent à un véhicule spécifique dans son topic I dédié.
// func (e *Emergency) publishIntentToVehicleTarget(vehicleID string, intent Intent) {
// 	topic := fmt.Sprintf("Anki/Vehicles/U/%s/I/%s", vehicleID, e.id) // Construction du topic spécifique au véhicule
// 	bs, _ := json.Marshal(intent)                                    // Sérialisation de l'intent en JSON
// 	token := e.client.Publish(topic, e.qos, false, bs)               // Publication du message
// 	token.WaitTimeout(3 * time.Second)                               // Attente de la confirmation de publication
// 	if token.Error() != nil {                                        // Gestion des erreurs
// 		log.Printf("publishIntentToVehicleTarget error: %v", token.Error())
// 	}
// }

// publishIntentToHosts relaie un Intent aux hôtes (Anki/Hosts/U/I/<callerID>)
// func (e *Emergency) publishIntentToHosts(intent Intent) {
// 	topic := fmt.Sprintf("Anki/Hosts/U/I/%s", e.id)    // Construction du topic pour les hôtes
// 	bs, _ := json.Marshal(intent)                      // Sérialisation de l'intent en JSON
// 	token := e.client.Publish(topic, e.qos, false, bs) // Publication du message
// 	token.WaitTimeout(3 * time.Second)                 // Attente de la confirmation de publication
// 	if token.Error() != nil {                          // Gestion des erreurs
// 		log.Printf("publishIntentToHosts error: %v", token.Error())
// 	}
// }

// handleStopMessage : gestionnaire de messages pour Emergency/U/E/stop
func (e *Emergency) handleStopMessage(client mqtt.Client, msg mqtt.Message) {
	var p stopMessagePayload                 // structure pour stocker la valeur du message d'arrêt
	err := json.Unmarshal(msg.Payload(), &p) // Tentative de désérialisation du message JSON
	// if err != nil {                          // Si la désérialisation échoue
	// 	// Tolère un simple payload "true"/"false" ou {"value": true}
	// 	s := strings.TrimSpace(strings.ToLower(string(msg.Payload())))
	// 	if s == "true" || s == `{"value":true}` {
	// 		p.Value = true // Valeur d'arrêt activée
	// 		err = nil      // Réinitialisation de l'erreur
	// 	} else if s == "false" || s == `{"value":false}` {
	// 		p.Value = false // Valeur d'arrêt désactivée
	// 		err = nil       // Réinitialisation de l'erreur
	// 	} else {
	// 		log.Printf("Emergency: couldn't parse stop payload: %s (err: %v)", string(msg.Payload()), err)
	// 		return
	// 	}
	// }
	prev := e.isStop()                                               // Récupération de l'état précédent d'arrêt d'urgence
	e.setStop(p.Value)                                               // Mise à jour de l'état d'arrêt d'urgence
	log.Printf("Emergency: STOP set to %v (prev=%v)", p.Value, prev) // Journalisation du changement d'état

	// Si le mode d'arrêt est activé, publier immédiatement un intent speed=0 à tous les véhicules
	if p.Value {
		intent := Intent{
			Type: "speed",
			Payload: map[string]interface{}{
				"velocity":     0,
				"acceleration": 0,
			},
		}
		log.Println("Emergency: publishing immediate speed=0 to all vehicles")
		e.publishIntentToVehicles(intent)
	}
	// si false, nous reprenons simplement le relais normal (aucune autre action nécessaire ici)
}

// listenKeyboardForStop écoute les entrées clavier pour déclencher l'arrêt d'urgence
func listenKeyboardForStop(e *Emergency) {
	for {
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "s" { // Si l'utilisateur tape 's' (pour stop)
			log.Println("STOP triggered via keyboard")
			e.setStop(true)

			intent := Intent{
				Type: "speed",
				Payload: map[string]interface{}{
					"velocity":     0,
					"acceleration": 0,
				},
			}
			e.publishIntentToVehicles(intent)
		}
	}
}

// ------------------------------------------------------------------------------------------
// Voir à partir d'ici pour le relais des messages RemoteControl
// ------------------------------------------------------------------------------------------

// mapRemoteTopicToMediate crée le topic mediate pour un topic RemoteControl donné
func mapRemoteTopicToMediate(remoteTopic string) string {
	return "Emergency/U/E/mediate/" + remoteTopic
}

// handleRemoteVehicleMessage sert de gestionnaire pour les messages RemoteControl véhicules
// Il relaie les messages sauf si stopActive. Il mappe simplement les topics RemoteControl vers Emergency/U/E/mediate/...
func (e *Emergency) handleRemoteVehicleMessage(client mqtt.Client, msg mqtt.Message) {
	if e.isStop() {
		log.Printf("Emergency: STOP active, ignoring remote message on %s", msg.Topic())
		return
	}

	mediateTopic := mapRemoteTopicToMediate(msg.Topic())
	token := client.Publish(mediateTopic, 1, false, msg.Payload())
	token.Wait()

	log.Printf("Emergency: forwarded %s -> %s", msg.Topic(), mediateTopic)
}

// Fonction principale qui permet de configurer le client MQTT, de s'abonner aux topics nécessaires et de gérer la boucle principale.
func main() {
	flag.Parse()

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

	// Souscrire au topic d'arrêt d'urgence
	stopTopic := "Emergency/U/E/stop"
	if tok := client.Subscribe(stopTopic, qos, em.handleStopMessage); tok.Wait() && tok.Error() != nil {
		log.Fatalf("Subscribe to stop topic failed: %v", tok.Error())
	}
	log.Printf("Subscribed to %s (retained-stop)", stopTopic)

	// Souscrire aux événements des véhicules RemoteControl
	remoteVehiclesTopic := "RemoteControl/+/E/vehicles/#"
	if tok := client.Subscribe(remoteVehiclesTopic, qos, em.handleRemoteVehicleMessage); tok.Wait() && tok.Error() != nil {
		log.Fatalf("Subscribe to remote vehicles failed: %v", tok.Error())
	}
	log.Printf("Subscribed to %s", remoteVehiclesTopic)

	// Optionnellement s'abonner RemoteControl hosts/discover
	// remoteHostsDiscover := "RemoteControl/+/E/hosts/discover"
	// if tok := client.Subscribe(remoteHostsDiscover, qos, em.handleRemoteHostDiscover); tok.Wait() && tok.Error() != nil {
	// 	log.Printf("Warning: subscribe hosts/discover failed: %v", tok.Error())
	// } else {
	// 	log.Printf("Subscribed to %s", remoteHostsDiscover)
	// }

	go listenKeyboardForStop(em) // Lancer l'écoute du clavier dans une goroutine séparée

	// boucle principale : continuer à fonctionner. Si STOP est actif, nous pourrions republier des impulsions périodiques de speed=0 si désiré.
	// Nous publierons un renforcement périodique du stop tant que stopActive pour garantir que les véhicules le reçoivent.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if em.isStop() {
				intent := Intent{
					Type: "speed",
					Payload: map[string]interface{}{
						"velocity":     0,
						"acceleration": 0,
					},
				}
				em.publishIntentToVehicles(intent)
			}
		}
	}
}
