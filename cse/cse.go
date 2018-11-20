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
	"strconv"
	"strings"
	"time"
)

func printLastError() {
	fmt.Printf("Error code %d: %s\n", int(C.cse_last_error_code()), C.GoString(C.cse_last_error_message()))
}

func createExecution() (execution *C.cse_execution) {
	execution = C.cse_execution_create(0.0, 0.1)
	return execution
}

func getSimulationTime(execution *C.cse_execution) (time float64) {
	var status C.cse_execution_status
	C.cse_execution_get_status(execution, &status)
	time = float64(status.current_time)
	return
}

func createLocalSlave(fmuPath string) (slave *C.cse_slave) {
	slave = C.cse_local_slave_create(C.CString(fmuPath))
	return
}

func createObserver() (observer *C.cse_observer) {
	observer = C.cse_membuffer_observer_create()
	return
}

func executionAddObserver(execution *C.cse_execution, observer *C.cse_observer) (observerIndex C.int) {
	observerIndex = C.cse_execution_add_observer(execution, observer)
	return
}

func executionAddSlave(execution *C.cse_execution, slave *C.cse_slave) int {
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

func executionDestroy(execution *C.cse_execution) {
	C.cse_execution_destroy(execution)
}

func executionStop(execution *C.cse_execution) {
	C.cse_execution_stop(execution)
}

func observerGetReals(observer *C.cse_observer, fmu structs.FMU) (realSignals []structs.Signal) {
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
		C.cse_observer_slave_get_real(observer, C.int(fmu.ExecutionIndex), &realValueRefs[0], C.ulonglong(numReals), &realOutVal[0])

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

func observerGetIntegers(observer *C.cse_observer, fmu structs.FMU) (intSignals []structs.Signal) {
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
		C.cse_observer_slave_get_integer(observer, C.int(fmu.ExecutionIndex), &intValueRefs[0], C.ulonglong(numIntegers), &intOutVal[0])

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

func observerGetRealSamples(observer *C.cse_observer, metaData structs.MetaData, signal *structs.TrendSignal, spec structs.TrendSpec) {
	fmu := findFmu(metaData, signal.Module)
	slaveIndex := C.int(fmu.ExecutionIndex)
	variableIndex := C.uint(findVariableIndex(fmu, signal.Signal, signal.Causality, signal.Type))

	stepNumbers := make([]C.longlong, 2)
	if spec.Auto {
		C.cse_observer_get_step_numbers_for_duration(observer, slaveIndex, C.double(spec.Range), &stepNumbers[0])
	} else {
		C.cse_observer_get_step_numbers(observer, slaveIndex, C.double(spec.Begin), C.double(spec.End), &stepNumbers[0])
	}

	first := stepNumbers[0]
	last := stepNumbers[1]

	numSamples := int(last) - int(first) + 1

	//fmt.Println("num samples:", numSamples, " first: ", int(first))

	cnSamples := C.ulonglong(numSamples)
	realOutVal := make([]C.double, numSamples)
	timeVal := make([]C.double, numSamples)
	timeStamps := make([]C.longlong, numSamples)
	actualNumSamples := C.cse_observer_slave_get_real_samples(observer, slaveIndex, variableIndex, first, cnSamples, &realOutVal[0], &timeStamps[0], &timeVal[0])

	trendVals := make([]float64, int(actualNumSamples))
	times := make([]float64, int(actualNumSamples))
	for i := 0; i < int(actualNumSamples); i++ {
		trendVals[i] = float64(realOutVal[i])
		times[i] = float64(timeVal[i])
	}
	signal.TrendTimestamps = times
	signal.TrendValues = trendVals
}

func findVariableIndex(fmu structs.FMU, signalName string, causality string, valueType string) (index int) {
	for _, variable := range fmu.Variables {
		if variable.Name == signalName && variable.Type == valueType && variable.Causality == causality {
			index = variable.ValueReference
		}
	}
	return
}

func TrendLoop(sim *Simulation, status *structs.SimulationStatus) {
	for {
		if len(status.TrendSignals) > 0 {
			var trend = &status.TrendSignals[0]
			switch trend.Type {
			case "Real":
				observerGetRealSamples(sim.Observer, sim.MetaData, trend, status.TrendSpec)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func parseFloat(argument string) float64 {
	f, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		log.Fatal(err)
		return 0.0
	}
	return f
}

func CommandLoop(sim *Simulation, command chan []string, status *structs.SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			switch cmd[0] {
			case "load":
				initializeSimulation(sim, cmd[1])
				status.Loaded = true
				status.ConfigDir = cmd[1]
				status.Status = "pause"
			case "teardown":
				status.Loaded = false
				status.Status = "stopped"
				status.ConfigDir = ""
				status.TrendSignals = []structs.TrendSignal{}
				status.Module = structs.Module{}
				executionDestroy(sim.Execution)
				sim = &Simulation{}
			case "stop":
				return
			case "pause":
				executionStop(sim.Execution)
				status.Status = "pause"
			case "play":
				executionStart(sim.Execution)
				status.Status = "play"
			case "trend":
				status.TrendSignals = append(status.TrendSignals, structs.TrendSignal{cmd[1], cmd[2], cmd[3], cmd[4], nil, nil})
			case "untrend":
				status.TrendSignals = []structs.TrendSignal{}
			case "trend-zoom":
				status.TrendSpec = structs.TrendSpec{Auto: false, Begin: parseFloat(cmd[1]), End: parseFloat(cmd[2])}
			case "trend-zoom-reset":
				status.TrendSpec = structs.TrendSpec{Auto: true, Range: parseFloat(cmd[1])}
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

func findFmu(metaData structs.MetaData, moduleName string) (foundFmu structs.FMU) {
	for _, fmu := range metaData.FMUs {
		if fmu.Name == moduleName {
			foundFmu = fmu
		}
	}
	return
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
				realSignals := observerGetReals(observer, fmu)
				intSignals := observerGetIntegers(observer, fmu)
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

func StateUpdateLoop(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *Simulation) {

	for {
		if simulationStatus.Loaded {
			state <- structs.JsonResponse{
				SimulationTime: getSimulationTime(sim.Execution),
				Modules:        getModuleNames(&sim.MetaData),
				Module:         getModuleData(simulationStatus, &sim.MetaData, sim.Observer),
				Loaded:         simulationStatus.Loaded,
				Status:         simulationStatus.Status,
				ConfigDir:      simulationStatus.ConfigDir,
				TrendSignals:   simulationStatus.TrendSignals,
			}
		} else {
			state <- structs.JsonResponse{
				Loaded: simulationStatus.Loaded,
				Status: simulationStatus.Status,
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, observer *C.cse_observer, metaData *structs.MetaData, fmuPath string) {
	log.Println("Loading: " + fmuPath)
	localSlave := createLocalSlave(fmuPath)
	fmu := metadata.ReadModelDescription(fmuPath)

	fmu.ExecutionIndex = executionAddSlave(execution, localSlave)
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

type Simulation struct {
	Execution *C.cse_execution
	Observer  *C.cse_observer
	MetaData  structs.MetaData
}

func CreateEmptySimulation() Simulation {
	return Simulation{}
}

func initializeSimulation(sim *Simulation, fmuDir string) {
	execution := createExecution()
	observer := createObserver()
	executionAddObserver(execution, observer)

	metaData := structs.MetaData{
		FMUs: []structs.FMU{},
	}
	paths := getFmuPaths(fmuDir)
	for _, path := range paths {
		addFmu(execution, observer, &metaData, path)
	}

	sim.Execution = execution
	sim.Observer = observer
	sim.MetaData = metaData
}
