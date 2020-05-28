// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"cosim-demo-app/libcosim"
	"cosim-demo-app/server"
	"cosim-demo-app/structs"
)

func main() {
	libcosim.SetupLogging()
	sim := libcosim.CreateEmptySimulation()

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan structs.JsonResponse, 10)

	simulationStatus := structs.SimulationStatus{
		Loaded:          false,
		Status:          "stopped",
		Trends:          []structs.Trend{},
		LibcosimVersion: libcosim.Version().LibcVer,
	}

	// Passing the channel to the go routine
	go libcosim.StateUpdateLoop(state, &simulationStatus, &sim)
	go libcosim.CommandLoop(state, &sim, cmd, &simulationStatus)
	go libcosim.TrendLoop(&sim, &simulationStatus)

	//Passing the channel to the server
	server.Server(cmd, state, &simulationStatus, &sim)
	close(cmd)
}
