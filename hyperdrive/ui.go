package hyperdrive

import (
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func carCard(client mqtt.Client, target string) fyne.CanvasObject {
	var isConnected bool = false
	var lightPayload = LightPayload{}

	// Data bindings for sliders
	velocityBinding := binding.NewFloat()
	velocityBinding.Set(0)
	accelerationBinding := binding.NewFloat()
	accelerationBinding.Set(0)

	lightStartBinding := binding.NewFloat()
	lightStartBinding.Set(0)
	lightEndBinding := binding.NewFloat()
	lightEndBinding.Set(0)
	lightFreqBinding := binding.NewFloat()
	lightFreqBinding.Set(0)

	// --- Connection ---
	var connectButton *widget.Button
	connectButton = widget.NewButton("Connect", func() {
		isConnected = !isConnected
		if isConnected {
			connectButton.SetText("Disconnect")
			connect(client, true, target)
		} else {
			connectButton.SetText("Connect")
			connect(client, false, target)
		}
	})

	// --- Movement ---
	velocitySlider := widget.NewSliderWithData(-100, 1000, velocityBinding)
	velocityValueLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(velocityBinding, "%.0f"))

	accelerationSlider := widget.NewSliderWithData(0, 2000, accelerationBinding)
	accelerationValueLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(accelerationBinding, "%.0f"))

	// Use a form layout for clean label/widget pairs
	movementForm := container.New(layout.NewFormLayout(),
		widget.NewLabel("Velocity:"),
		// Use a border layout to put the value label next to the slider
		container.NewBorder(nil, nil, nil, velocityValueLabel, velocitySlider),
		widget.NewLabel("Acceleration:"),
		container.NewBorder(nil, nil, nil, accelerationValueLabel, accelerationSlider),
	)

	speedApplyButton := widget.NewButton("Apply", func() {
		v, _ := velocityBinding.Get()
		a, _ := accelerationBinding.Get()
		err := speed(client, float32(v), float32(a), target)
		if err != nil {
			log.Println("[UI] Could not send speed payload correctly:", err)
		}
	})

	// --- Lane Change ---
	laneChangeLabel := widget.NewLabel("Lane Change: ")
	laneChangeLeftButton := widget.NewButton("<<", func() {
		// TODO: Implement lane change left logic
		log.Println("[UI] Lane change left for", target)
	})
	laneChangeRightButton := widget.NewButton(">>", func() {
		// TODO: Implement lane change right logic
		log.Println("[UI] Lane change right for", target)
	})
	laneCancelButton := widget.NewButton("Cancel", func() {
		// TODO: Implement lane change cancel logic
		log.Println("[UI] Lane change cancel for", target)
	})
	// Use a spacer to push buttons to the right
	laneChangeBox := container.NewHBox(laneChangeLabel, layout.NewSpacer(), laneChangeLeftButton, laneChangeRightButton, layout.NewSpacer(), laneCancelButton)

	// --- Lights ---
	lightTypeOptions := []string{"Front Green", "Front Red", "Tail", "Engine Red", "Engine Green", "Engine Blue"}
	lightTypeSelect := widget.NewSelect(lightTypeOptions, nil)
	lightTypeSelect.PlaceHolder = "Select Light..."

	lightEffectOptions := []string{"Off", "Steady", "Fade", "Pulse", "Flash", "Strobe"}
	lightEffectSelect := widget.NewSelect(lightEffectOptions, nil)
	lightEffectSelect.PlaceHolder = "Select Effect..."

	// Sliders are used for SpinBox equivalent
	lightStartSlider := widget.NewSliderWithData(0, 15, lightStartBinding)
	lightStartValueLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(lightStartBinding, "%.0f"))

	lightEndSlider := widget.NewSliderWithData(0, 15, lightEndBinding)
	lightEndValueLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(lightEndBinding, "%.0f"))

	lightFreqSlider := widget.NewSliderWithData(0, 255, lightFreqBinding)
	lightFreqValueLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(lightFreqBinding, "%.0f"))

	lightsForm := container.New(layout.NewFormLayout(),
		widget.NewLabel("Light:"), lightTypeSelect,
		widget.NewLabel("Effect:"), lightEffectSelect,
		widget.NewLabel("Start:"), container.NewBorder(nil, nil, nil, lightStartValueLabel, lightStartSlider),
		widget.NewLabel("End:"), container.NewBorder(nil, nil, nil, lightEndValueLabel, lightEndSlider),
		widget.NewLabel("Frequency:"), container.NewBorder(nil, nil, nil, lightFreqValueLabel, lightFreqSlider),
	)

	lightApplyBtn := widget.NewButton("Apply", func() {
		// Ensure dropdowns have a selection
		if lightTypeSelect.Selected == "" || lightEffectSelect.Selected == "" {
			log.Println("[UI] Please select a light and an effect.")
			return
		}

		start, _ := lightStartBinding.Get()
		end, _ := lightEndBinding.Get()
		freq, _ := lightFreqBinding.Get()

		effect := LightEffect{
			Effect:    strings.ToLower(lightEffectSelect.Selected),
			Start:     int(start), // Assuming LightEffect uses int
			End:       int(end),   // Assuming LightEffect uses int
			Frequency: int(freq),  // Assuming LightEffect uses int
		}

		selected := lightTypeSelect.Selected
		switch selected {
		case "Front Green":
			lightPayload.FrontGreen = effect
		case "Front Red":
			lightPayload.FrontRed = effect
		case "Tail":
			lightPayload.Tail = effect
		case "Engine Red":
			lightPayload.EngineRed = effect
		case "Engine Green":
			lightPayload.EngineGreen = effect
		case "Engine Blue":
			lightPayload.EngineBlue = effect
		}

		err := lights(client, lightPayload, target)
		if err != nil {
			log.Println("[UI] Could not send lights payload correctly:", err)
		}
	})

	// --- Assemble Card ---
	cardContent := container.NewVBox(
		container.NewCenter(connectButton),
		widget.NewSeparator(),
		movementForm,
		container.NewCenter(speedApplyButton),
		widget.NewSeparator(),
		laneChangeBox,
		widget.NewSeparator(),
		lightsForm,
		container.NewCenter(lightApplyBtn),
	)

	// Use a Card for each car, which is Fyne's equivalent of a QGroupBox
	return widget.NewCard(target, "", cardContent)
}

func initialPrompt(window fyne.Window, client mqtt.Client) fyne.CanvasObject {

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
			log.Println(hostIntentTopicEntry.Text, "\n", vehicleIntentTopicFormatEntry, "\n", hostDiscoverVehicleTopicEntry)

			err := SyncSubscription(client, "discoverSubscription", hostIntentTopicEntry.Text, DiscoverTopic, true)
			if err != nil {
				log.Fatal("Could not sync with the discover subscription.")
			}

			// Wait half a second to be sure that the subscription went trhough.
			time.Sleep(500 * time.Millisecond)

			vehicleList, err := InitializeRemote(client, hostDiscoverVehicleTopicEntry.Text, vehicleIntentTopicFormatEntry.Text)
			if err != nil {
				log.Fatal("Could not initialize the remote:", err)
			}

			// Build a list of card widgets, one for each car
			carCards := []fyne.CanvasObject{}
			for _, car := range vehicleList {
				carCards = append(carCards, carCard(client, car))
			}

			// Place all car cards in a VBox, which is then put in a VScroll
			content := container.NewVScroll(
				container.NewVBox(carCards...),
			)

			// replace the form by the cars
			window.SetContent(content)

			// disconnect from the cars if the app is closed
			window.SetOnClosed(func() {
				for _, car := range vehicleList {
					connect(client, false, car)
				}
			})
		},
	}

	return form
}

// App is the main Fyne application entry point.
// This replaces the original App() function.
func App(client mqtt.Client) {
	client.Publish("HelloWorld", 1, false, "Hello, world !")

	a := app.New()
	w := a.NewWindow("Hyperdrive RemoteControl")

	// First, show a form where the user has to insert the different topics
	// This makes it decoupled (?)
	form := initialPrompt(w, client)
	w.SetContent(form)
	w.Resize(fyne.NewSize(450, 700))
	w.ShowAndRun()
}
