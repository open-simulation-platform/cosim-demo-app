// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"cse-server-go/cse"
	"cse-server-go/server"
	"cse-server-go/structs"
)

func main() {
	cse.SetupLogging()
	sim := cse.CreateEmptySimulation()

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan structs.JsonResponse, 10)

	simulationStatus := structs.SimulationStatus{
		Loaded: false,
		Status: "stopped",
		Trends: []structs.Trend{},
	}

	// Passing the channel to the go routine
	go cse.StateUpdateLoop(state, &simulationStatus, &sim)
	go cse.CommandLoop(state, &sim, cmd, &simulationStatus)
	go cse.TrendLoop(&sim, &simulationStatus)

	//Passing the channel to the server
	server.Server(cmd, state, &simulationStatus, &sim)
	close(cmd)
}
