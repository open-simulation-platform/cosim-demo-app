package cse

/*
	#include <cse.h>
*/
import "C"
import (
	"cse-server-go/metadata"
	"cse-server-go/structs"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func printLastError() {
	fmt.Printf("Error code %d: %s\n", int(C.cse_last_error_code()), C.GoString(C.cse_last_error_message()))
}

func CreateExecution() (execution *C.cse_execution) {
	execution = C.cse_execution_create(0.0, 0.01)
	return execution
}

func CreateLocalSlave(fmuPath string) (slave *C.cse_slave) {
	slave = C.cse_local_slave_create(C.CString(fmuPath))
	return
}

func CreateObserver() (observer *C.cse_observer) {
	observer = C.cse_membuffer_observer_create()
	return
}

func ExecutionAddObserver(execution *C.cse_execution, observer *C.cse_observer) (observerIndex C.int) {
	observerIndex = C.cse_execution_add_observer(execution, observer)
	return
}

func ObserverAddSlave(observer *C.cse_observer, slave *C.cse_slave) int {
	slaveIndex := C.cse_observer_add_slave(observer, slave)
	if slaveIndex < 0 {
		printLastError()
		//C.cse_observer_destroy(observer)
	}
	return int(slaveIndex)
}

func ExecutionAddSlave(execution *C.cse_execution, slave *C.cse_slave) int {
	slaveIndex := C.cse_execution_add_slave(execution, slave)
	if slaveIndex < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	return int(slaveIndex)
}

func executionStart(execution *C.cse_execution) {
	C.cse_execution_start(execution)
}

func executionStop(execution *C.cse_execution) {
	C.cse_execution_stop(execution)
}

type executionStatus struct {
	current_time float64
}

func executionGetStatus(execution *C.cse_execution) (status executionStatus) {
	cStatus := C.cse_execution_status{}
	C.cse_execution_get_status(execution, &cStatus)
	status.current_time = float64(cStatus.current_time)
	return
}

func ObserverGetReals(observer *C.cse_observer, fmu structs.FMU) (realSignals []structs.Signal) {
	var realValueRefs []C.uint
	var realVariables []structs.Variable
	var numReals int
	for _, variable := range fmu.Variables {
		if variable.Type == "Real" {
			ref := C.uint(variable.ValueReference)
			realValueRefs = append(realValueRefs, ref)
			realVariables = append(realVariables, variable)
			numReals++
		}
	}

	if numReals > 0 {
		realOutVal := make([]C.double, numReals)
		C.cse_observer_slave_get_real(observer, C.int(fmu.ObserverIndex), &realValueRefs[0], C.ulonglong(numReals), &realOutVal[0])

		realSignals = make([]structs.Signal, numReals)
		for k := range realVariables {
			realSignals[k] = structs.Signal{
				Name:      realVariables[k].Name,
				Causality: realVariables[k].Causality,
				Type:      realVariables[k].Type,
				Value:     float64(realOutVal[k]),
			}
		}
	}
	return realSignals
}

func ObserverGetIntegers(observer *C.cse_observer, fmu structs.FMU) (intSignals []structs.Signal) {
	var intValueRefs []C.uint
	var intVariables []structs.Variable
	var numIntegers int
	for _, variable := range fmu.Variables {
		if variable.Type == "Integer" {
			ref := C.uint(variable.ValueReference)
			intValueRefs = append(intValueRefs, ref)
			intVariables = append(intVariables, variable)
			numIntegers++
		}
	}

	if numIntegers > 0 {
		intOutVal := make([]C.int, numIntegers)
		C.cse_observer_slave_get_integer(observer, C.int(fmu.ObserverIndex), &intValueRefs[0], C.ulonglong(numIntegers), &intOutVal[0])

		intSignals = make([]structs.Signal, numIntegers)
		for k := range intVariables {
			intSignals[k] = structs.Signal{
				Name:      intVariables[k].Name,
				Causality: intVariables[k].Causality,
				Type:      intVariables[k].Type,
				Value:     int(intOutVal[k]),
			}
		}
	}
	return intSignals
}

func ObserverGetRealSamples(observer *C.cse_observer, nSamples int, signal *structs.TrendSignal) {
	fromSample := 0
	if len(signal.TrendTimestamps) > 0 {
		fromSample = signal.TrendTimestamps[len(signal.TrendTimestamps)-1]
	}
	slaveIndex := C.int(0)
	variableIndex := C.uint(0)
	cnSamples := C.ulonglong(nSamples)
	realOutVal := make([]C.double, nSamples)
	timeStamps := make([]C.long, nSamples)
	actualNumSamples := C.cse_observer_slave_get_real_samples(observer, slaveIndex, variableIndex, C.long(fromSample), cnSamples, &realOutVal[0], &timeStamps[0])

	for i := 0; i < int(actualNumSamples); i++ {
		signal.TrendTimestamps = append(signal.TrendTimestamps, int(timeStamps[i]))
		signal.TrendValues = append(signal.TrendValues, float64(realOutVal[i]))
	}

}

func Polling(sim SimulatorBeta, status *structs.SimulationStatus) {
	for {
		if len(status.TrendSignals) > 0 {
			ObserverGetRealSamples(sim.Observer, 10, &status.TrendSignals[0])
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func Simulate(sim SimulatorBeta, command chan []string, status *structs.SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			switch cmd[0] {
			case "stop":
				return
			case "pause":
				executionStop(sim.Execution)
				status.Status = "pause"
			case "play":
				executionStart(sim.Execution)
				status.Status = "play"
			case "trend":
				status.TrendSignals = append(status.TrendSignals, structs.TrendSignal{cmd[1], cmd[2], nil, nil})
			case "untrend":
				status.TrendSignals = []structs.TrendSignal{}
			case "module":
				status.Module = structs.Module{
					Name: cmd[1],
				}
			default:
				fmt.Println("Empty command, mildt sagt not good: ", cmd)
			}
		}
	}
}

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
				realSignals := ObserverGetReals(observer, fmu)
				intSignals := ObserverGetIntegers(observer, fmu)
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

func StatePoll(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim SimulatorBeta) {

	for {
		state <- structs.JsonResponse{
			Modules:      getModuleNames(&sim.MetaData),
			Module:       getModuleData(simulationStatus, &sim.MetaData, sim.Observer),
			Status:       simulationStatus.Status,
			TrendSignals: simulationStatus.TrendSignals,
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func AddFmu(execution *C.cse_execution, observer *C.cse_observer, metaData *structs.MetaData, fmuPath string) {
	log.Println("Loading: " + fmuPath)
	localSlave := CreateLocalSlave(fmuPath)
	fmu := metadata.ReadModelDescription(fmuPath)

	fmu.ExecutionIndex = ExecutionAddSlave(execution, localSlave)
	fmu.ObserverIndex = ObserverAddSlave(observer, localSlave)
	metaData.FMUs = append(metaData.FMUs, fmu)
}

func GetFmuPaths(loadFolder string) (paths []string) {
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

type SimulatorBeta struct {
	Execution *C.cse_execution
	Observer  *C.cse_observer
	MetaData  structs.MetaData
}

func CreateSimulation() SimulatorBeta {
	execution := CreateExecution()
	observer := CreateObserver()
	ExecutionAddObserver(execution, observer)

	dataDir := os.Getenv("TEST_DATA_DIR")
	localSlave := CreateLocalSlave(dataDir + "/fmi2/Clock.fmu")
	fmu := metadata.ReadModelDescription(dataDir + "/fmi2/Clock.fmu")

	slaveExecutionIndex := ExecutionAddSlave(execution, localSlave)
	fmu.ExecutionIndex = slaveExecutionIndex
	observerSlaveIndex := ObserverAddSlave(observer, localSlave)
	fmu.ObserverIndex = observerSlaveIndex
	metaData := structs.MetaData{
		FMUs: []structs.FMU{fmu},
	}
	return SimulatorBeta{
		Execution: execution,
		Observer:  observer,
		MetaData:  metaData,
	}
}
