# Hyperdrive remote control

This repository implements the last mandatory project developped in the Process Control course of the university of Fribourg.

## Running

To install the project, first install go, clone the repository, and then run either `main.go` to run the _RemoteControl_ app, or `emergency/main.go` to run the _Emergency_ app.

```sh
# install go
sudo apt install golang-go # ubuntu
sudo dnf install golang # fedora
winget install GoLang.Go # windows
brew install go # macos

# clone the repository
git clone https://github.com/Uhrbaan/hyperdrive-remote.git
cd hyperdrive-remote

# run the apps
go run main.go # RemoteControl app
go run emergency/main.go # Emergency app
```

<!-- The first time you run an app, it will install all dependencies which might take some time.
If you encounter problems during the installation of `fyne` (a dependency used to create the graphical applications), please refer to <https://docs.fyne.io/started/quick/>. -->

> Please note that for the apps to work, both must be running at the same time, and you _must_ me connected to the hyperdrive wifi. Also make sure that the topics provided at the app startup are correct, although they should be if you haven't changed the default setup.

## Hyperdrive Remote

A remote-control application for Anki Overdrive cars, featuring MQTT-based communication, pathfinding, lane changing, and a graphical user interface. This project is designed for research and educational purposes, enabling advanced control and automation of Overdrive vehicles.

## Features

- **MQTT Remote Control:**
  - Discover and connect to one or multiple cars via MQTT.
  - Control car speed, lane changes, and lights remotely.
- **Pathfinding & Lane Change:**
  - Advanced pathfinding algorithms for automated driving.
  - Lane change logic for overtaking and track navigation.
- **Graphical User Interface:**
  - Built with [Fyne](https://fyne.io/) for cross-platform desktop control.
- **Track & Vehicle Modeling:**
  - YAML and Graphviz-based track definitions for flexible layouts.
  - Vehicle and track abstractions for simulation and planning.
- **Emergency Controls:**
  - Emergency stop and safety features.

## Project Structure

```
assets/           # Track definitions (YAML, Graphviz)
emergency/        # Emergency stop and safety logic
hyperdrive/       # Core remote control logic (connect, drive, lights, UI)
pathfind/         # Pathfinding, lane change, and track/vehicle modeling
main.go           # Application entry point
go.mod, go.sum    # Go module dependencies
```

## Getting Started

### Prerequisites

- Go 1.24+
- Anki Overdrive cars and track
- MQTT broker (e.g., Mosquitto)

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/hyperdrive-remote.git
   cd hyperdrive-remote
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```

### Running the Application

1. Start your MQTT broker (default: `10.42.0.1:1883`).
2. Run the application:
   ```sh
   go run main.go
   ```
3. Use the GUI to discover, connect, and control cars.

### Track Configuration

- Edit YAML files in `assets/` to define your track layout.
- Use Graphviz files for visualizing and debugging track graphs.

## Dependencies

- [Fyne](https://fyne.io/) (GUI)
- [Eclipse Paho MQTT](https://github.com/eclipse/paho.mqtt.golang) (MQTT client)
- [Google UUID](https://github.com/google/uuid)
- [Dominik Braun Graph](https://github.com/dominikbraun/graph) (Graph algorithms)
- [Go-YAML](https://github.com/goccy/go-yaml)
