package cse

/*
	#include <cse.h>
*/
import "C"
import (
	"cse-server-go/metadata"
	"cse-server-go/structs"
	"fmt"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"log"
	"math"
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
	startTime := C.cse_time_point(0.0 * 1e9)
	stepSize := C.cse_duration(0.1 * 1e9)
	execution = C.cse_execution_create(startTime, stepSize)
	return execution
}

func createSsdExecution(ssdDir string) (execution *C.cse_execution) {
	startTime := C.cse_time_point(0.0 * 1e9)
	execution = C.cse_ssp_execution_create(C.CString(ssdDir), startTime)
	return execution
}

type executionStatus struct {
	time                 float64
	realTimeFactor       float64
	isRealTimeSimulation bool
}

func getExecutionStatus(execution *C.cse_execution) (execStatus executionStatus) {
	var status C.cse_execution_status
	C.cse_execution_get_status(execution, &status)
	nanoTime := int64(status.current_time)
	execStatus.time = float64(nanoTime) * 1e-9
	execStatus.realTimeFactor = float64(status.real_time_factor)
	execStatus.isRealTimeSimulation = int(status.is_real_time_simulation) > 0
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

func createFileObserver(logPath string) (observer *C.cse_observer) {
	observer = C.cse_file_observer_create(C.CString(logPath))
	return
}

func executionAddObserver(execution *C.cse_execution, observer *C.cse_observer) (observerIndex C.cse_observer_index) {
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

func observerDestroy(observer *C.cse_observer) {
	C.cse_observer_destroy(observer)
}

func executionStop(execution *C.cse_execution) {
	C.cse_execution_stop(execution)
}

func executionEnableRealTime(execution *C.cse_execution) {
	C.cse_execution_enable_real_time_simulation(execution)
}

func executionDisableRealTime(execution *C.cse_execution) {
	C.cse_execution_disable_real_time_simulation(execution)
}

func uglyNanFix(value C.double) interface{} {
	floatValue := float64(value)
	if math.IsNaN(floatValue) {
		return "NaN"
	}
	return floatValue
}

func observerGetReals(observer *C.cse_observer, fmu structs.FMU) (realSignals []structs.Signal) {
	var realValueRefs []C.cse_variable_index
	var realVariables []structs.Variable
	var numReals int
	for _, variable := range fmu.Variables {
		if variable.Type == "Real" {
			ref := C.cse_variable_index(variable.ValueReference)
			realValueRefs = append(realValueRefs, ref)
			realVariables = append(realVariables, variable)
			numReals++
		}
	}

	if numReals > 0 {
		realOutVal := make([]C.double, numReals)
		C.cse_observer_slave_get_real(observer, C.cse_slave_index(fmu.ExecutionIndex), &realValueRefs[0], C.size_t(numReals), &realOutVal[0])

		realSignals = make([]structs.Signal, numReals)
		for k := range realVariables {
			realSignals[k] = structs.Signal{
				Name:      realVariables[k].Name,
				Causality: realVariables[k].Causality,
				Type:      realVariables[k].Type,
				Value:     uglyNanFix(realOutVal[k]),
			}
		}
	}
	return realSignals
}

func observerGetIntegers(observer *C.cse_observer, fmu structs.FMU) (intSignals []structs.Signal) {
	var intValueRefs []C.cse_variable_index
	var intVariables []structs.Variable
	var numIntegers int
	for _, variable := range fmu.Variables {
		if variable.Type == "Integer" {
			ref := C.cse_variable_index(variable.ValueReference)
			intValueRefs = append(intValueRefs, ref)
			intVariables = append(intVariables, variable)
			numIntegers++
		}
	}

	if numIntegers > 0 {
		intOutVal := make([]C.int, numIntegers)
		C.cse_observer_slave_get_integer(observer, C.cse_slave_index(fmu.ExecutionIndex), &intValueRefs[0], C.size_t(numIntegers), &intOutVal[0])

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

func observerGetRealSamples(observer *C.cse_observer, metaData *structs.MetaData, signal *structs.TrendSignal, spec structs.TrendSpec) {
	fmu := findFmu(metaData, signal.Module)
	slaveIndex := C.cse_slave_index(fmu.ExecutionIndex)
	variableIndex := C.cse_variable_index(findVariableIndex(fmu, signal.Signal, signal.Causality, signal.Type))

	stepNumbers := make([]C.cse_step_number, 2)
	if spec.Auto {
		duration := C.cse_duration(spec.Range * 1e9)
		C.cse_observer_get_step_numbers_for_duration(observer, slaveIndex, duration, &stepNumbers[0])
	} else {
		tBegin := C.cse_time_point(spec.Begin * 1e9)
		tEnd := C.cse_time_point(spec.End * 1e9)
		C.cse_observer_get_step_numbers(observer, slaveIndex, tBegin, tEnd, &stepNumbers[0])
	}

	first := stepNumbers[0]
	last := stepNumbers[1]

	numSamples := int(last) - int(first) + 1
	cnSamples := C.size_t(numSamples)
	realOutVal := make([]C.double, numSamples)
	timeVal := make([]C.cse_time_point, numSamples)
	timeStamps := make([]C.cse_step_number, numSamples)
	actualNumSamples := C.cse_observer_slave_get_real_samples(observer, slaveIndex, variableIndex, first, cnSamples, &realOutVal[0], &timeStamps[0], &timeVal[0])

	trendVals := make([]float64, int(actualNumSamples))
	times := make([]float64, int(actualNumSamples))
	for i := 0; i < int(actualNumSamples); i++ {
		trendVals[i] = float64(realOutVal[i])
		times[i] = 1e-9 * float64(timeVal[i])
	}
	signal.TrendTimestamps = times
	signal.TrendValues = trendVals
}

func setReal(execution *C.cse_execution, slaveIndex int, variableIndex int, value float64) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.double, 1)
	v[0] = C.double(value)
	C.cse_execution_slave_set_real(execution, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
}

func setInteger(execution *C.cse_execution, slaveIndex int, variableIndex int, value int) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.int, 1)
	v[0] = C.int(value)
	C.cse_execution_slave_set_integer(execution, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
}

func setVariableValue(sim *Simulation, module string, signal string, causality string, valueType string, value string) {
	fmu := findFmu(sim.MetaData, module)
	varIndex := findVariableIndex(fmu, signal, causality, valueType)
	switch valueType {
	case "Real":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println(err)
		} else {
			setReal(sim.Execution, fmu.ExecutionIndex, varIndex, val)
		}
	case "Integer":
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Println(err)
		} else {
			setInteger(sim.Execution, fmu.ExecutionIndex, varIndex, val)
		}
	default:
		fmt.Println("Can't set this value:", value)
	}
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
		status.Mutex.Lock()
		if len(status.TrendSignals) > 0 {
			for i, _ := range status.TrendSignals {
				var trend = &status.TrendSignals[i]
				switch trend.Type {
				case "Real":
					observerGetRealSamples(sim.Observer, sim.MetaData, trend, status.TrendSpec)
				}
			}
		}
		status.Mutex.Unlock()
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

func simulationTeardown(sim *Simulation) {
	executionDestroy(sim.Execution)
	observerDestroy(sim.Observer)
	if nil != sim.FileObserver {
		observerDestroy(sim.FileObserver)
	}
	sim.Execution = nil
	sim.Observer = nil
	sim.FileObserver = nil
	sim.MetaData = &structs.MetaData{}
}

func initializeSimulation(sim *Simulation, fmuDir string, logDir string) {
	metaData := structs.MetaData{
		FMUs: []structs.FMU{},
	}
	var execution *C.cse_execution
	if hasSsdFile(fmuDir) {
		execution = createSsdExecution(fmuDir)
		addSsdMetadata(execution, &metaData, fmuDir)
	} else {
		execution = createExecution()
		paths := getFmuPaths(fmuDir)
		for _, path := range paths {
			addFmu(execution, &metaData, path)
		}
	}

	observer := createObserver()
	executionAddObserver(execution, observer)

	var fileObserver *C.cse_observer
	if len(logDir) > 0 {
		fileObserver := createFileObserver(logDir)
		executionAddObserver(execution, fileObserver)
	}

	sim.Execution = execution
	sim.Observer = observer
	sim.FileObserver = fileObserver
	sim.MetaData = &metaData
}

func CommandLoop(sim *Simulation, command chan []string, status *structs.SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			status.Mutex.Lock()
			switch cmd[0] {
			case "load":
				initializeSimulation(sim, cmd[1], cmd[2])
				status.Loaded = true
				status.ConfigDir = cmd[1]
				status.Status = "pause"
			case "teardown":
				status.Loaded = false
				status.Status = "stopped"
				status.ConfigDir = ""
				status.TrendSignals = []structs.TrendSignal{}
				status.Module = structs.Module{}
				simulationTeardown(sim)
			case "stop":
				return
			case "pause":
				executionStop(sim.Execution)
				status.Status = "pause"
			case "play":
				executionStart(sim.Execution)
				status.Status = "play"
			case "enable-realtime":
				executionEnableRealTime(sim.Execution)
			case "disable-realtime":
				executionDisableRealTime(sim.Execution)
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
			case "set-value":
				setVariableValue(sim, cmd[1], cmd[2], cmd[3], cmd[4], cmd[5])
			default:
				fmt.Println("Unknown command, this is not good: ", cmd)
			}
			status.Mutex.Unlock()
		}
	}
}

func findFmu(metaData *structs.MetaData, moduleName string) (foundFmu structs.FMU) {
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

func getModuleData(observer *C.cse_observer, fmu structs.FMU) (module structs.Module) {
	realSignals := observerGetReals(observer, fmu)
	intSignals := observerGetIntegers(observer, fmu)
	var signals []structs.Signal
	signals = append(signals, realSignals...)
	signals = append(signals, intSignals...)
	module.Name = fmu.Name
	module.Signals = signals
	return
}

func findModuleData(status *structs.SimulationStatus, metaData *structs.MetaData, observer *C.cse_observer) (module structs.Module) {
	if len(status.Module.Name) > 0 {
		for _, fmu := range metaData.FMUs {
			if fmu.Name == status.Module.Name {
				module = getModuleData(observer, fmu)
			}
		}
	}
	return
}

func GetSignalValue(module string, cardinality string, signal string) int {
	return 1
}

func SimulationStatus(simulationStatus *structs.SimulationStatus, sim *Simulation) structs.JsonResponse {
	simulationStatus.Mutex.Lock()
	defer simulationStatus.Mutex.Unlock()
	virtualMemoryStats, _ := mem.VirtualMemory()
	if simulationStatus.Loaded {
		execStatus := getExecutionStatus(sim.Execution)
		return structs.JsonResponse{
			SimulationTime:       execStatus.time,
			RealTimeFactor:       execStatus.realTimeFactor,
			IsRealTimeSimulation: execStatus.isRealTimeSimulation,
			Modules:              getModuleNames(sim.MetaData),
			Module:               findModuleData(simulationStatus, sim.MetaData, sim.Observer),
			Loaded:               simulationStatus.Loaded,
			Status:               simulationStatus.Status,
			ConfigDir:            simulationStatus.ConfigDir,
			TrendSignals:         simulationStatus.TrendSignals,
			Memory:               virtualMemoryStats,
		}
	} else {
		return structs.JsonResponse{
			Loaded: simulationStatus.Loaded,
			Status: simulationStatus.Status,
		}
	}
}

func StateUpdateLoop(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *Simulation) {

	for {
		state <- SimulationStatus(simulationStatus, sim)
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, metaData *structs.MetaData, fmuPath string) {
	log.Println("Loading: " + fmuPath)
	localSlave := createLocalSlave(fmuPath)
	fmu := metadata.ReadModelDescription(fmuPath)

	fmu.ExecutionIndex = executionAddSlave(execution, localSlave)
	metaData.FMUs = append(metaData.FMUs, fmu)
}

func addFmuSsd(metaData *structs.MetaData, name string, index int, fmuPath string) {
	log.Println("Parsing: " + fmuPath)
	fmu := metadata.ReadModelDescription(fmuPath)
	fmu.Name = name
	fmu.ExecutionIndex = index
	metaData.FMUs = append(metaData.FMUs, fmu)
}

func addSsdMetadata(execution *C.cse_execution, metaData *structs.MetaData, fmuDir string) {
	nSlaves := C.cse_execution_get_num_slaves(execution)
	var slaveInfos = make([]C.cse_slave_info, int(nSlaves))
	C.cse_execution_get_slave_infos(execution, &slaveInfos[0], nSlaves)
	for _, info := range slaveInfos {
		name := C.GoString(&info.name[0])
		source := C.GoString(&info.source[0])
		index := int(info.index)
		pathToFmu := filepath.Join(fmuDir, source)
		addFmuSsd(metaData, name, index, pathToFmu)
	}
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

func hasSsdFile(loadFolder string) bool {
	info, e := os.Stat(loadFolder)
	if os.IsNotExist(e) {
		fmt.Println("Load folder does not exist!")
		return false
	} else if !info.IsDir() {
		fmt.Println("Load folder is not a directory!")
		return false
	} else {
		files, err := ioutil.ReadDir(loadFolder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if f.Name() == "SystemStructure.ssd" {
				return true
			}
		}
	}
	return false
}

type Simulation struct {
	Execution    *C.cse_execution
	Observer     *C.cse_observer
	FileObserver *C.cse_observer
	MetaData     *structs.MetaData
}

func CreateEmptySimulation() Simulation {
	return Simulation{}
}
