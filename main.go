package main

import "time"

func loop(state chan JsonResponse) {
	for {
		state <- JsonResponse{
			Status:      "running",
			SignalValue: lastOutValue,
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	execution := createExecution()
	slaveIndex := executionAddfmu(execution, "C:/dev/osp/cse-core/test/data/fmi2/Clock.fmu")

	// Creating a command channel
	cmd := make(chan string, 10)
	state := make(chan JsonResponse, 10)

	// Passing the channel to the go routine
	go simulate(execution, slaveIndex, cmd)
	go loop(state)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}
