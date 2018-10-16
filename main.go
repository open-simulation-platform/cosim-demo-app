package main

/*
	#include <cse.h>
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getModuleNames(metaData *MetaData) []string {
	nModules := len(metaData.FMUs)
	modules := make([]string, nModules)
	for i := range metaData.FMUs {
		modules[i] = metaData.FMUs[i].Name
	}
	return modules
}

func getModuleData(status *SimulationStatus, metaData *MetaData, observer *C.cse_observer) (module Module) {
	if len(status.Module.Name) > 0 {
		for _, fmu := range metaData.FMUs {
			if fmu.Name == status.Module.Name {
				realSignals := observerGetReals(observer, fmu)
				intSignals := observerGetIntegers(observer, fmu)
				var signals []Signal
				signals = append(signals, realSignals...)
				signals = append(signals, intSignals...)
				module.Name = fmu.Name
				module.Signals = signals
			}
		}
	}

	return module
}

func statePoll(state chan JsonResponse, simulationStatus *SimulationStatus, metaData *MetaData, observer *C.cse_observer) {
	for {
		state <- JsonResponse{
			Modules:      getModuleNames(metaData),
			Module:       getModuleData(simulationStatus, metaData, observer),
			Status:       simulationStatus.Status,
			TrendSignals: simulationStatus.TrendSignals,
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, observer *C.cse_observer, metaData *MetaData, fmuPath string) {
	log.Println("Loading: " + fmuPath)
	localSlave := createLocalSlave(fmuPath)
	fmu := ReadModelDescription(fmuPath)

	fmu.ExecutionIndex = executionAddSlave(execution, localSlave)
	fmu.ObserverIndex = observerAddSlave(observer, localSlave)
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
	execution := createExecution()
	observer := createObserver()
	executionAddObserver(execution, observer)

	metaData := &MetaData{
		FMUs: []FMU{},
	}
	dataDir := os.Getenv("TEST_DATA_DIR")
	paths := getFmuPaths(dataDir + "/fmi2")
	for _, path := range paths {
		addFmu(execution, observer, metaData, path)
	}

	// Creating a command channel
	cmd := make(chan []string, 10)
	state := make(chan JsonResponse, 10)

	simulationStatus := &SimulationStatus{
		Status:       "pause",
		TrendSignals: []TrendSignal{},
	}

	// Passing the channel to the go routine
	go statePoll(state, simulationStatus, metaData, observer)
	go simulate(execution, cmd, simulationStatus)
	go polling(observer, simulationStatus)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}
