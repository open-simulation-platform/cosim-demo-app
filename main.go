package main

import "time"

func loop(state chan JsonResponse, simulationStatus *SimulationStatus) {
	for {
		state <- JsonResponse{
			Modules: []string{"Clock"},
			Module: Module{
				Signals: []Signal{
					{
						Name:  "Clock",
						Value: lastOutValue,
					},
				},
				Name: "Clock",
			},
			Status: simulationStatus.Status,
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	execution := createExecution()
	localSlave := createLocalSlave("C:/dev/osp/cse-core/test/data/fmi2/Clock.fmu")
	observer := createObserver()
	executionAddSlave(execution, localSlave)
	observerAddSlave(observer, localSlave)
	executionAddObserver(execution, observer)

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan JsonResponse, 10)

	simulationStatus := &SimulationStatus{
		Status: "pause",
		TrendSignals: []TrendSignal{},
	}

	// Passing the channel to the go routine
	go loop(state, simulationStatus)
	go simulate(execution, observer, cmd, simulationStatus)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}