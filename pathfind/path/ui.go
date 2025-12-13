package path

import (
	"encoding/json"
	"fmt"
	"hyperdrive/remote/pathfind/util"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var gridSections = [5][5]int{
	{13, 20, 04, 21, 16},
	{1, 07, 5, 8, 2},
	{9, 11, 6, 10, 12},
	{18, 22, 0, 23, 19}, // important note: 17 changed to 0.
	{14, 24, 3, 25, 15},
}

const (
	vehicleTargetTopic = "/hobHq10yb9dKwxrdfhtT/vehicle/target"
)

type tilePayload struct {
	ID int `json:"id"`
}

func UI(client mqtt.Client) {
	// go randomPositions(client)

	// Start the application
	a := app.New()
	w := a.NewWindow("Visual Car Section Tracker")
	w.Resize(fyne.NewSize(600, 500))

	absolute := map[int]*fyne.Animation{} // Used to store rectangle references according to the ID
	preditction := map[int]*fyne.Animation{}
	cells := []fyne.CanvasObject{}
	var previousTarget *canvas.Rectangle = nil
	for _, row := range gridSections {
		for _, col := range row {
			// image
			image := canvas.NewImageFromFile(fmt.Sprintf("assets/ID_%02d.png", col))
			image.FillMode = canvas.ImageFillContain

			// rectangle (to color the grid cell)
			targetRect := canvas.NewRectangle(color.RGBA{100, 100, 0, 20})
			targetRect.Hide()
			rect := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
			animation := canvas.NewColorRGBAAnimation(color.RGBA{255, 0, 0, 200}, color.RGBA{0, 0, 0, 0}, time.Second*2, func(c color.Color) {
				rect.FillColor = c
				canvas.Refresh(rect)
			})

			rect2 := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
			animation2 := canvas.NewColorRGBAAnimation(color.RGBA{0, 255, 0, 200}, color.RGBA{0, 0, 0, 0}, time.Second*2, func(c color.Color) {
				rect2.FillColor = c
				canvas.Refresh(rect)
			})

			absolute[col] = animation
			preditction[col] = animation2

			// button, don't put one for the crossing, since it is not allowed to stop there.
			if col != 0 {
				button := widget.NewButton("", func() {
					if previousTarget != nil {
						previousTarget.Hide()
					}
					targetRect.Show()
					previousTarget = targetRect

					payload, err := json.Marshal(tilePayload{col})
					if err != nil {
						return
					}
					client.Publish(vehicleTargetTopic, 1, false, payload)
				})
				cells = append(cells, container.New(layout.NewStackLayout(), button, image, rect, rect2, targetRect))
			} else {
				cells = append(cells, container.New(layout.NewStackLayout(), image, rect, rect2, targetRect))
			}

		}
	}

	client.Subscribe(vehicleAbsolutePositionTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		fmt.Println("Received a received an absolute position.")
		var data tilePayload
		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			return
		}

		if _, ok := absolute[data.ID]; ok {
			absolute[data.ID].Start()
		}
	})

	client.Subscribe(vehiclePredictionTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		fmt.Println("Received a prediction.")
		var data tilePayload
		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			return
		}

		if _, ok := preditction[data.ID]; ok {
			preditction[data.ID].Start()
		}
	})

	grid := container.New(layout.NewGridLayout(5), cells...)

	vehicleIdEntry := widget.NewEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Car ID:",
				Widget: vehicleIdEntry,
			},
		},
		OnSubmit: func() {
			payload, _ := json.Marshal(util.VehicleIdPayload{vehicleIdEntry.Text})
			client.Publish(util.VehicleIDTopic, 1, false, payload)
			w.SetContent(grid)
		},
	}
	w.SetContent(form)
	w.ShowAndRun()
}
