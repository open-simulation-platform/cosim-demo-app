package main

import (
	"cse-server-go/cse"
	"cse-server-go/server"
	"cse-server-go/structs"
)

func main() {
	sim := cse.CreateEmptySimulation()

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan structs.JsonResponse, 10)

	simulationStatus := structs.SimulationStatus{
		Loaded:       false,
		Status:       "stopped",
		TrendSignals: []structs.TrendSignal{},
		TrendSpec: structs.TrendSpec{
			Auto:  true,
			Range: 10.0},
	}

	// Passing the channel to the go routine
	go cse.StateUpdateLoop(state, &simulationStatus, &sim)
	go cse.CommandLoop(&sim, cmd, &simulationStatus)
	go cse.TrendLoop(&sim, &simulationStatus)

	//Passing the channel to the server
	server.Server(cmd, state, &simulationStatus, &sim)
	close(cmd)
}
