package main

import "time"

func statePoll(state chan JsonResponse, simulationStatus *SimulationStatus) {
	for {
		state <- JsonResponse{
			Modules: []string{"Clock", "Cock"},
			Module: Module{
				Signals: []Signal{
					{
						Name:  simulationStatus.SelectedModule,
						Value: lastOutValue,
					},
				},
				Name: "Clock",
			},
			Status:       simulationStatus.Status,
			TrendSignals: simulationStatus.TrendSignals,
		}
		time.Sleep(5000 * time.Millisecond)
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
		TrendSignals: []TrendSignal{
			{
				Module:          "Clock",
				Signal:          "Clock",
				TrendValues:     []float64{},
				TrendTimestamps: []int{},
			},
		},
	}

	// Passing the channel to the go routine
	go statePoll(state, simulationStatus)
	go simulate(execution, cmd, simulationStatus)
	go polling(observer, simulationStatus)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}
