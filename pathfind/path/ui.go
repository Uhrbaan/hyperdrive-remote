package path

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"strings"
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
	vehicleIDTopic     = "/hobHq10yb9dKwxrdfhtT/vehicle/id"
	// fullPathTopic      = "/hobHq10yb9dKwxrdfhtT/graph/fullPath" // Topic to receive the full path (Kev)
)

type tilePayload struct {
	ID int `json:"id"`
}

type vehicleIdPayload struct {
	ID string `json:"id"`
}

// NOUVEAU : Structure pour recevoir le chemin complet
// type fullPathPayload struct {
// 	Path []string `json:"path"` // Path est une liste de nœuds (ex: "13.curve.outer")
// }

// NOUVEAU : Fonction utilitaire pour extraire l'ID de tuile à partir du nom du nœud
func getTileIDFromNode(node string) (int, error) {
	parts := strings.Split(node, ".")
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid node format: %s", node)
	}
	// L'ID de tuile est les deux premiers caractères (ex: "13")
	return strconv.Atoi(parts[0])
}

func UI(client mqtt.Client) {
	// go randomPositions(client)

	// Start the application
	a := app.New()
	w := a.NewWindow("Visual Car Section Tracker")
	w.Resize(fyne.NewSize(600, 500))

	absolute := map[int]*fyne.Animation{} // Used to store rectangle references according to the ID
	preditction := map[int]*fyne.Animation{}

	// --- NOUVEAU : Map pour la visualisation du chemin (Bleu) ---
	pathVisualization := map[int]*canvas.Rectangle{}
	var currentPathTiles []int // Pour garder une trace des tuiles bleues

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

			// --- NOUVEAU : Rectangle Bleu pour le Chemin ---
			rect3 := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
			pathColor := color.RGBA{0, 0, 255, 128} // Bleu semi-transparent
			rect3.FillColor = pathColor
			rect3.Hide()

			pathVisualization[col] = rect3
			// ---------------------------------------------

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
				cells = append(cells, container.New(layout.NewStackLayout(), button, image, rect, rect2, rect3, targetRect))
			} else {
				cells = append(cells, container.New(layout.NewStackLayout(), image, rect, rect2, rect3, targetRect))
			}

		}
	}

	client.Subscribe(vehiclePositionTopic, 1, func(c mqtt.Client, m mqtt.Message) {
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

	// --- NOUVEAU : Abonnement pour le Chemin Complet (Bleu) ---
	client.Subscribe(fullPathTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		fmt.Println("Received a full path for visualization.")
		var data fullPathPayload
		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			fmt.Printf("Error unmarshalling full path payload: %v\n", err)
			return
		}

		// 1. Cacher le chemin précédent
		for _, tileID := range currentPathTiles {
			if rect, ok := pathVisualization[tileID]; ok {
				rect.Hide()
			}
		}
		currentPathTiles = nil // Réinitialiser la liste

		// 2. Afficher le nouveau chemin
		newPathTiles := make(map[int]bool) // Utiliser une map pour dédoublonner les tuiles (un chemin peut passer plusieurs fois sur un ID de tuile)
		for _, node := range data.Path {
			tileID, err := getTileIDFromNode(node)
			if err != nil {
				fmt.Printf("Warning: Could not get tile ID from node '%s': %v\n", node, err)
				continue
			}

			// Si la tuile est dans notre grille
			if _, ok := pathVisualization[tileID]; ok {
				newPathTiles[tileID] = true
			}
		}

		// 3. Mettre à jour l'UI pour les nouvelles tuiles du chemin
		for tileID := range newPathTiles {
			if rect, ok := pathVisualization[tileID]; ok {
				rect.Show()
				currentPathTiles = append(currentPathTiles, tileID)
			}
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
			payload, _ := json.Marshal(vehicleIdPayload{vehicleIdEntry.Text})
			client.Publish(vehicleIDTopic, 1, false, payload)
			w.SetContent(grid)
		},
	}
	w.SetContent(form)
	w.ShowAndRun()
}
