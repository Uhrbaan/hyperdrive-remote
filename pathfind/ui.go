package main

import (
	"encoding/json"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var gridSections = [5][5]string{
	{"ID_13", "ID_20", "ID_04", "ID_21", "ID_16"},
	{"ID_01", "ID_07", "ID_05", "ID_08", "ID_02"},
	{"ID_09", "ID_11", "ID_06", "ID_10", "ID_12"},
	{"ID_18", "ID_22", "ID_17", "ID_23", "ID_19"},
	{"ID_14", "ID_24", "ID_03", "ID_25", "ID_15"},
}

const (
	vehicleTargetTopic = "/hobHq10yb9dKwxrdfhtT/vehicle/target"
)

// func randomPositions(client mqtt.Client) {
// 	for {
// 		id := gridSections[rand.IntN(len(gridSections))][rand.IntN(len(gridSections[0]))]
// 		payload, _ := json.Marshal(positionPayload{id})
// 		client.Publish(vehiclePositionTopic, 1, false, payload)
// 		time.Sleep(2 * time.Second)
// 	}
// }

func UI(client mqtt.Client) {
	// go randomPositions(client)

	// Start the application
	a := app.New()
	w := a.NewWindow("Visual Car Section Tracker")
	w.Resize(fyne.NewSize(600, 500))

	rectangles := map[string]*canvas.Rectangle{} // Used to store rectangle references according to the ID
	cells := []fyne.CanvasObject{}
	var previousTarget *canvas.Rectangle = nil
	for _, row := range gridSections {
		for _, col := range row {
			// image
			image := canvas.NewImageFromFile("assets/" + col + ".png")
			image.FillMode = canvas.ImageFillContain

			// rectangle (to color the grid cell)
			targetRect := canvas.NewRectangle(color.RGBA{100, 100, 0, 20})
			targetRect.Hide()
			rect := canvas.NewRectangle(color.RGBA{100, 0, 0, 20})
			rect.Hide()
			rectangles[col] = rect

			// button
			button := widget.NewButton("", func() {
				if previousTarget != nil {
					previousTarget.Hide()
				}
				targetRect.Show()
				previousTarget = targetRect

				payload, err := json.Marshal(positionPayload{col})
				if err != nil {
					return
				}
				client.Publish(vehicleTargetTopic, 1, false, payload)
			})

			cells = append(cells, container.New(layout.NewStackLayout(), button, image, rect, targetRect))
		}
	}

	// previousVehiclePosition := ""
	// client.Subscribe(vehiclePositionTopic, 1, func(c mqtt.Client, m mqtt.Message) {
	// 	var data positionPayload
	// 	err := json.Unmarshal(m.Payload(), &data)
	// 	if err != nil {
	// 		return
	// 	}
	// 	rectangles[data.ID].Show()
	// 	if previousVehiclePosition != "" {
	// 		rectangles[previousVehiclePosition].Hide()
	// 	}
	// 	previousVehiclePosition = data.ID
	// })

	grid := container.New(layout.NewGridLayout(5), cells...)
	w.SetContent(grid)
	w.ShowAndRun()
}
