#!/bin/bash

set -xe

# 1. Lancement de l'application RemoteControl (contrôle manuel)
go run . &

# 2. Lancement de l'application Emergency (médiation et sécurité)
go run ./emergency/main.go &

# 3. Lancement de l'application Pathfinding (suivi et navigation)
go run ./pathfind/main.go &

echo "Launched all three programs: RemoteControl, Emergency, and Pathfinding."