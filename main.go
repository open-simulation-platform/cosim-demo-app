package main

import (
	"cse-server-go/cse"
	"cse-server-go/server"
	"cse-server-go/structs"
	"os"
)

func main() {
	execution := cse.CreateExecution()
	observer := cse.CreateObserver()
	cse.ExecutionAddObserver(execution, observer)

	metaData := &structs.MetaData{
		FMUs: []structs.FMU{},
	}
	dataDir := os.Getenv("TEST_DATA_DIR")
	paths := cse.GetFmuPaths(dataDir + "/fmi2")
	for _, path := range paths {
		cse.AddFmu(execution, observer, metaData, path)
	}

	sim := cse.CreateSimulation()

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan structs.JsonResponse, 10)

	simulationStatus := &structs.SimulationStatus{
		Status:       "pause",
		TrendSignals: []structs.TrendSignal{},
	}

	// Passing the channel to the go routine
	go cse.StatePoll(state, simulationStatus, sim)
	go cse.Simulate(sim, cmd, simulationStatus)
	go cse.Polling(sim, simulationStatus)

	//Passing the channel to the server
	server.Server(cmd, state)
	close(cmd)
}
