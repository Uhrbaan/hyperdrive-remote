#!/bin/bash

set -xe

go run . &
go run ./emergency/main.go &

echo "Launched both programs."