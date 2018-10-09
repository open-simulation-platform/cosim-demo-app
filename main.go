package main

/*
	#include <cse.h>
*/
import "C"
import (
	"os"
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
		for i := range metaData.FMUs {
			if metaData.FMUs[i].Name == status.Module.Name {
				fmu := metaData.FMUs[i]
				realSignals := observerGetReals(observer, fmu)
				var signals []Signal
				signals = append(module.Signals, realSignals...)
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
			Modules: getModuleNames(metaData),
			Module:  getModuleData(simulationStatus, metaData, observer),
			Status:  simulationStatus.Status,
			//TrendSignals: simulationStatus.TrendSignals,
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	execution := createExecution()
	observer := createObserver()
	executionAddObserver(execution, observer)

	dataDir := os.Getenv("TEST_DATA_DIR")
	localSlave := createLocalSlave(dataDir + "/fmi2/Current.fmu")
	fmu := ReadModelDescription(dataDir + "/fmi2/Current.fmu")

	slaveExecutionIndex := executionAddSlave(execution, localSlave)
	fmu.ExecutionIndex = slaveExecutionIndex
	observerSlaveIndex := observerAddSlave(observer, localSlave)
	fmu.ObserverIndex = observerSlaveIndex

	metaData := &MetaData{
		FMUs: []FMU{fmu},
	}

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
	go statePoll(state, simulationStatus, metaData, observer)
	go simulate(execution, cmd, simulationStatus)
	go polling(observer, simulationStatus)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}
