package main

/*
	#include <cse.h>
*/
import "C"
import (
	"cse-server-go/cse"
	"cse-server-go/metadata"
	"cse-server-go/server"
	"cse-server-go/structs"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getModuleNames(metaData *structs.MetaData) []string {
	nModules := len(metaData.FMUs)
	modules := make([]string, nModules)
	for i := range metaData.FMUs {
		modules[i] = metaData.FMUs[i].Name
	}
	return modules
}

func getModuleData(status *structs.SimulationStatus, metaData *structs.MetaData, observer *C.cse_observer) (module structs.Module) {
	if len(status.Module.Name) > 0 {
		for _, fmu := range metaData.FMUs {
			if fmu.Name == status.Module.Name {
				realSignals := cse.ObserverGetReals(observer, fmu)
				intSignals := cse.ObserverGetIntegers(observer, fmu)
				var signals []structs.Signal
				signals = append(signals, realSignals...)
				signals = append(signals, intSignals...)
				module.Name = fmu.Name
				module.Signals = signals
			}
		}
	}

	return module
}

func statePoll(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, metaData *structs.MetaData, observer *C.cse_observer) {

	for {
		state <- structs.JsonResponse{
			Modules:      getModuleNames(metaData),
			Module:       getModuleData(simulationStatus, metaData, observer),
			Status:       simulationStatus.Status,
			TrendSignals: simulationStatus.TrendSignals,
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, observer *C.cse_observer, metaData *structs.MetaData, fmuPath string) {
	log.Println("Loading: " + fmuPath)
	localSlave := cse.CreateLocalSlave(fmuPath)
	fmu := metadata.ReadModelDescription(fmuPath)

	fmu.ExecutionIndex = cse.ExecutionAddSlave(execution, localSlave)
	fmu.ObserverIndex = cse.ObserverAddSlave(observer, localSlave)
	metaData.FMUs = append(metaData.FMUs, fmu)
}

func getFmuPaths(loadFolder string) (paths []string) {
	info, e := os.Stat(loadFolder)
	if os.IsNotExist(e) {
		fmt.Println("Load folder does not exist!")
		return
	} else if !info.IsDir() {
		fmt.Println("Load folder is not a directory!")
		return
	} else {
		files, err := ioutil.ReadDir(loadFolder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".fmu") {
				paths = append(paths, filepath.Join(loadFolder, f.Name()))
			}
		}
	}
	return paths
}

func main() {
	execution := cse.CreateExecution()
	observer := cse.CreateObserver()
	cse.ExecutionAddObserver(execution, observer)

	metaData := &structs.MetaData{
		FMUs: []structs.FMU{},
	}
	dataDir := os.Getenv("TEST_DATA_DIR")
	paths := getFmuPaths(dataDir + "/fmi2")
	for _, path := range paths {
		addFmu(execution, observer, metaData, path)
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
	go statePoll(state, simulationStatus, &sim.MetaData, sim.Observer)
	go cse.Simulate(sim.Execution, cmd, simulationStatus)
	go cse.Polling(sim.Observer, simulationStatus)

	//Passing the channel to the server
	server.Server(cmd, state)
	close(cmd)
}
