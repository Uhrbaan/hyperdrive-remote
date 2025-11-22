# Hyperdrive remote control 
This repository implements the last mandatory project developped in the Process Control course of the university of Fribourg. 

## Running 
To install the project, first install go, clone the repository, and then run either `main.go` to run the *RemoteControl* app, or `emergency/main.go` to run the *Emergency* app.

```sh 
# install go 
sudo apt install golang-go # ubuntu
sudo dnf install golang # fedora

# clone the repository
git clone https://github.com/Uhrbaan/hyperdrive-remote.git
cd hyperdrive-remote

# run the apps
go run main.go # RemoteControl app 
go run emergency/main.go # Emergency app
```

The first time you run an app, it will install all dependencies which might take some time.

> Please note that for the apps to work, both must be running at the same time, and you *must* me connected to the hyperdrive wifi. Also make sure that the topics provided at the app startup are correct, although they should be if you haven't changed the default setup.