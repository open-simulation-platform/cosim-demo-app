package cse

/*
	#include <cse.h>
*/
import "C"
import (
	"cse-server-go/metadata"
	"cse-server-go/structs"
	"errors"
	"fmt"
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

func createLocalSlave(fmuPath string) *C.cse_slave {
	return C.cse_local_slave_create(C.CString(fmuPath))
}

func createObserver() (observer *C.cse_observer) {
	observer = C.cse_buffered_membuffer_observer_create(C.size_t(1))
	return
}

func createTrendObserver() (observer *C.cse_observer) {
	observer = C.cse_time_series_observer_create()
	return
}

func createFileObserver(logPath string) (observer *C.cse_observer) {
	observer = C.cse_file_observer_create(C.CString(logPath))
	return
}

func executionAddObserver(execution *C.cse_execution, observer *C.cse_observer) {
	C.cse_execution_add_observer(execution, observer)
}

func executionAddSlave(execution *C.cse_execution, slave *C.cse_slave) int {
	slaveIndex := C.cse_execution_add_slave(execution, slave)
	if slaveIndex < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	return int(slaveIndex)
}

func executionStart(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_start(execution)
	if int(success) < 0 {
		return false, "Unable to start simulation"
	} else {
		return true, "Simulation is running"
	}
}

func executionDestroy(execution *C.cse_execution) {
	C.cse_execution_destroy(execution)
}

func observerDestroy(observer *C.cse_observer) {
	C.cse_observer_destroy(observer)
}

func executionStop(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_stop(execution)
	if int(success) < 0 {
		return false, "Unable to stop simulation"
	} else {
		return true, "Simulation is paused"
	}
}

func executionEnableRealTime(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_enable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, "Unable to enable real time"
	} else {
		return true, "Real time execution enabled"
	}
}

func executionDisableRealTime(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_disable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, "Unable to disable real time"
	} else {
		return true, "Real time execution disabled"
	}
}

func uglyNanFix(value C.double) interface{} {
	floatValue := float64(value)
	if math.IsNaN(floatValue) {
		return "NaN"
	} else if math.IsInf(floatValue, 1) {
		return "+Inf"
	} else if math.IsInf(floatValue, -1) {
		return "-Inf"
	}
	return floatValue
}

func observerGetReals(observer *C.cse_observer, variables []structs.Variable, slaveIndex int) (realSignals []structs.Signal) {
	var realValueRefs []C.cse_variable_index
	var realVariables []structs.Variable
	var numReals int
	for _, variable := range variables {
		if variable.Type == "Real" {
			ref := C.cse_variable_index(variable.ValueReference)
			realValueRefs = append(realValueRefs, ref)
			realVariables = append(realVariables, variable)
			numReals++
		}
	}

	if numReals > 0 {
		realOutVal := make([]C.double, numReals)
		C.cse_observer_slave_get_real(observer, C.cse_slave_index(slaveIndex), &realValueRefs[0], C.size_t(numReals), &realOutVal[0])

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

func observerGetIntegers(observer *C.cse_observer, variables []structs.Variable, slaveIndex int) (intSignals []structs.Signal) {
	var intValueRefs []C.cse_variable_index
	var intVariables []structs.Variable
	var numIntegers int
	for _, variable := range variables {
		if variable.Type == "Integer" {
			ref := C.cse_variable_index(variable.ValueReference)
			intValueRefs = append(intValueRefs, ref)
			intVariables = append(intVariables, variable)
			numIntegers++
		}
	}

	if numIntegers > 0 {
		intOutVal := make([]C.int, numIntegers)
		C.cse_observer_slave_get_integer(observer, C.cse_slave_index(slaveIndex), &intValueRefs[0], C.size_t(numIntegers), &intOutVal[0])

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

func observerGetRealSamples(observer *C.cse_observer, signal *structs.TrendSignal, spec structs.TrendSpec) {
	slaveIndex := C.cse_slave_index(signal.SlaveIndex)
	variableIndex := C.cse_variable_index(signal.ValueReference)

	stepNumbers := make([]C.cse_step_number, 2)
	var success C.int
	if spec.Auto {
		duration := C.cse_duration(spec.Range * 1e9)
		success = C.cse_observer_get_step_numbers_for_duration(observer, slaveIndex, duration, &stepNumbers[0])
	} else {
		tBegin := C.cse_time_point(spec.Begin * 1e9)
		tEnd := C.cse_time_point(spec.End * 1e9)
		success = C.cse_observer_get_step_numbers(observer, slaveIndex, tBegin, tEnd, &stepNumbers[0])
	}
	if int(success) < 0 {
		return
	}
	first := stepNumbers[0]
	last := stepNumbers[1]

	numSamples := int(last) - int(first) + 1
	cnSamples := C.size_t(numSamples)
	realOutVal := make([]C.double, numSamples)
	timeVal := make([]C.cse_time_point, numSamples)
	timeStamps := make([]C.cse_step_number, numSamples)
	actualNumSamples := C.cse_observer_slave_get_real_samples(observer, slaveIndex, variableIndex, first, cnSamples, &realOutVal[0], &timeStamps[0], &timeVal[0])
	ns := int(actualNumSamples)
	if ns <= 0 {
		return
	}
	trendVals := make([]float64, ns)
	times := make([]float64, ns)
	for i := 0; i < ns; i++ {
		trendVals[i] = float64(realOutVal[i])
		times[i] = 1e-9 * float64(timeVal[i])
	}
	signal.TrendTimestamps = times
	signal.TrendValues = trendVals
}

func setReal(execution *C.cse_execution, slaveIndex int, variableIndex int, value float64) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.double, 1)
	v[0] = C.double(value)
	success := C.cse_execution_slave_set_real(execution, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, "Unable to set real variable value"
	} else {
		return true, "Successfully set real variable value"
	}
}

func setInteger(execution *C.cse_execution, slaveIndex int, variableIndex int, value int) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.int, 1)
	v[0] = C.int(value)
	success := C.cse_execution_slave_set_integer(execution, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, "Unable to set integer variable value"
	} else {
		return true, "Successfully set integer variable value"
	}
}

func setVariableValue(sim *Simulation, module string, signal string, causality string, valueType string, value string) (bool, string) {
	fmu := findFmu(sim.MetaData, module)
	varIndex := findVariableIndex(fmu, signal, causality, valueType)
	switch valueType {
	case "Real":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println(err)
			return false, err.Error()
		} else {
			return setReal(sim.Execution, fmu.ExecutionIndex, varIndex, val)
		}
	case "Integer":
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Println(err)
			return false, err.Error()
		} else {
			return setInteger(sim.Execution, fmu.ExecutionIndex, varIndex, val)
		}
	default:
		message := strCat("Can't set this value: ", value)
		fmt.Println(message)
		return false, message
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
		if len(status.TrendSignals) > 0 {
			for i, _ := range status.TrendSignals {
				var trend = &status.TrendSignals[i]
				switch trend.Type {
				case "Real":
					observerGetRealSamples(sim.TrendObserver, trend, status.TrendSpec)
				}
			}
		}
		time.Sleep(1000 * time.Millisecond)
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

func simulationTeardown(sim *Simulation) (bool, string) {
	executionDestroy(sim.Execution)
	observerDestroy(sim.Observer)
	observerDestroy(sim.TrendObserver)
	if nil != sim.FileObserver {
		observerDestroy(sim.FileObserver)
	}
	sim.Execution = nil
	sim.Observer = nil
	sim.TrendObserver = nil
	sim.FileObserver = nil
	sim.MetaData = &structs.MetaData{}
	return true, "Simulation teardown successful"
}

func validateConfigDir(fmuDir string) (bool, string) {
	if _, err := os.Stat(fmuDir); os.IsNotExist(err) {
		return false, strCat(fmuDir, " does not exist")
	}
	files, err := ioutil.ReadDir(fmuDir)
	if err != nil {
		return false, err.Error()
	}
	var hasFMUs = false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".fmu") {
			hasFMUs = true
		}
	}
	if !hasFMUs {
		return false, strCat(fmuDir, " does not contain any FMUs")
	}
	return true, ""
}

func initializeSimulation(sim *Simulation, fmuDir string, logDir string) (bool, string) {
	if valid, message := validateConfigDir(fmuDir); !valid {
		return false, message
	}
	metaData := structs.MetaData{
		FMUs: []structs.FMU{},
	}
	var execution *C.cse_execution
	if hasSsdFile(fmuDir) {
		execution = createSsdExecution(fmuDir)
		if execution == nil {
			return false, "Could not create execution from SystemStructure.ssd file"
		}
		addSsdMetadata(execution, &metaData, fmuDir)
	} else {
		execution = createExecution()
		if execution == nil {
			return false, "Could not create execution"
		}
		paths := getFmuPaths(fmuDir)
		for _, path := range paths {
			success := addFmu(execution, &metaData, path)
			if !success {
				return false, strCat("Could not add FMU to execution: ", path)
			}
		}
	}

	observer := createObserver()
	executionAddObserver(execution, observer)

	trendObserver := createTrendObserver()
	executionAddObserver(execution, trendObserver)

	var fileObserver *C.cse_observer
	if len(logDir) > 0 {
		fileObserver := createFileObserver(logDir)
		executionAddObserver(execution, fileObserver)
	}

	sim.Execution = execution
	sim.Observer = observer
	sim.TrendObserver = trendObserver
	sim.FileObserver = fileObserver
	sim.MetaData = &metaData
	return true, "Simulation loaded successfully"
}

func strCat(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

func toVariableType(valueType string) (C.cse_variable_type, error) {
	switch valueType {
	case "Real":
		return C.CSE_REAL, nil
	case "Integer":
		return C.CSE_INTEGER, nil
	case "Boolean":
		return C.CSE_BOOLEAN, nil
	case "String":
		return C.CSE_STRING, nil
	}
	return C.CSE_REAL, errors.New(strCat("Unknown variable type:", valueType))
}

func observerStartObserving(observer *C.cse_observer, slaveIndex int, valueType string, varIndex int) error {
	variableType, err := toVariableType(valueType)
	if err != nil {
		return err
	}
	C.cse_observer_start_observing(observer, C.cse_slave_index(slaveIndex), variableType, C.cse_variable_index(varIndex))
	return nil
}

func observerStopObserving(observer *C.cse_observer, slaveIndex int, valueType string, varIndex int) error {
	variableType, err := toVariableType(valueType)
	if err != nil {
		return err
	}
	C.cse_observer_stop_observing(observer, C.cse_slave_index(slaveIndex), variableType, C.cse_variable_index(varIndex))
	return nil
}

func addToTrend(sim *Simulation, status *structs.SimulationStatus, module string, signal string, causality string, valueType string, valueReference string, plotType string) (bool, string) {
	fmu := findFmu(sim.MetaData, module)
	varIndex, err := strconv.Atoi(valueReference)
	if err != nil {
		message := strCat("Cannot parse valueReference as integer ", valueReference, ", ", err.Error())
		log.Println(message)
		return false, message
	}
	err = observerStartObserving(sim.TrendObserver, fmu.ExecutionIndex, valueType, varIndex)
	if err != nil {
		message := strCat("Cannot start observing variable ", err.Error())
		log.Println(message)
		return false, message
	}
	status.TrendSignals = append(status.TrendSignals, structs.TrendSignal{
		Module:         module,
		SlaveIndex:     fmu.ExecutionIndex,
		Signal:         signal,
		Causality:      causality,
		Type:           valueType,
		PlotType:       plotType,
		ValueReference: varIndex})
	return true, "Added variable to trend"
}

func removeAllFromTrend(sim *Simulation, status *structs.SimulationStatus) (bool, string) {
	var success = true
	var message = "Removed all variables from trend"
	for _, trendSignal := range status.TrendSignals {
		err := observerStopObserving(sim.TrendObserver, trendSignal.SlaveIndex, trendSignal.Type, trendSignal.ValueReference)
		if err != nil {
			message = strCat("Cannot stop observing variable: ", err.Error())
			success = false
			log.Println("Cannot stop observing", err)
		}
	}
	status.TrendSignals = []structs.TrendSignal{}
	return success, message
}

func executeCommand(cmd []string, sim *Simulation, status *structs.SimulationStatus) (feedback structs.CommandFeedback) {
	var success = false
	var message = "No feedback implemented for this command"
	switch cmd[0] {
	case "load":
		success, message = initializeSimulation(sim, cmd[1], cmd[2])
		if success {
			status.Loaded = true
			status.ConfigDir = cmd[1]
			status.Status = "pause"
			status.MetaChan <- sim.MetaData
		}
	case "teardown":
		status.Loaded = false
		status.Status = "stopped"
		status.ConfigDir = ""
		status.TrendSignals = []structs.TrendSignal{}
		status.Module = ""
		success, message = simulationTeardown(sim)
		status.MetaChan <- sim.MetaData
	case "pause":
		success, message = executionStop(sim.Execution)
		status.Status = "pause"
	case "play":
		success, message = executionStart(sim.Execution)
		status.Status = "play"
	case "enable-realtime":
		success, message = executionEnableRealTime(sim.Execution)
	case "disable-realtime":
		success, message = executionDisableRealTime(sim.Execution)
	case "trend":
		success, message = addToTrend(sim, status, cmd[1], cmd[2], cmd[3], cmd[4], cmd[5], cmd[6])
	case "untrend":
		success, message = removeAllFromTrend(sim, status)
	case "trend-zoom":
		status.TrendSpec = structs.TrendSpec{Auto: false, Begin: parseFloat(cmd[1]), End: parseFloat(cmd[2])}
		success = true
		message = strCat("Trending values from ", cmd[1], " to ", cmd[2])
	case "trend-zoom-reset":
		status.TrendSpec = structs.TrendSpec{Auto: true, Range: parseFloat(cmd[1])}
		success = true
		message = strCat("Trending last ", cmd[1], " seconds")
	case "set-value":
		success, message = setVariableValue(sim, cmd[1], cmd[2], cmd[3], cmd[4], cmd[5])
	case "get-module-data":
		status.MetaChan <- sim.MetaData
		success = true
		message = "Fetched metadata"
	case "signals":
		success, message = setSignalSubscriptions(status, cmd)
	default:
		message = "Unknown command, this is not good"
		fmt.Println(message, cmd)
	}
	return structs.CommandFeedback{Success: success, Message: message, Command: cmd[0]}
}

func CommandLoop(state chan structs.JsonResponse, sim *Simulation, command chan []string, status *structs.SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			feedback := executeCommand(cmd, sim, status)
			state <- GenerateJsonResponse(status, sim, feedback)
		}
	}
}

func setSignalSubscriptions(status *structs.SimulationStatus, cmd []string) (bool, string) {
	var variables []structs.Variable
	var message = "Successfully set signal subscriptions"
	var success = true
	if len(cmd) > 1 {
		status.Module = cmd[1]
		for i, signal := range cmd {
			if i > 1 {
				parts := strings.Split(signal, ",")
				valRef, err := strconv.Atoi(parts[3])
				if err != nil {
					message = strCat("Could not parse value reference from: ", signal, ", ", err.Error())
					log.Println(message)
					success = false
				} else {
					variables = append(variables,
						structs.Variable{
							Name:           parts[0],
							Causality:      parts[1],
							Type:           parts[2],
							ValueReference: valRef,
						})
				}
			}
		}
	} else {
		message = "Successfully reset signal subscriptions"
	}
	status.SignalSubscriptions = variables
	return success, message
}

func findFmu(metaData *structs.MetaData, moduleName string) (foundFmu structs.FMU) {
	for _, fmu := range metaData.FMUs {
		if fmu.Name == moduleName {
			foundFmu = fmu
		}
	}
	return
}

func findModuleData(status *structs.SimulationStatus, metaData *structs.MetaData, observer *C.cse_observer) (module structs.Module) {
	if len(status.SignalSubscriptions) > 0 {

		slaveIndex := findFmu(metaData, status.Module).ExecutionIndex
		realSignals := observerGetReals(observer, status.SignalSubscriptions, slaveIndex)
		intSignals := observerGetIntegers(observer, status.SignalSubscriptions, slaveIndex)
		var signals []structs.Signal
		signals = append(signals, realSignals...)
		signals = append(signals, intSignals...)

		module.Signals = signals
		module.Name = status.Module
	}
	return
}

func GetSignalValue(module string, cardinality string, signal string) int {
	return 1
}

func maybeGetMetaData(metaChan <-chan *structs.MetaData) *structs.MetaData {
	select {
	case m := <-metaChan:
		return m
	default:
		return nil
	}
}

func GenerateJsonResponse(status *structs.SimulationStatus, sim *Simulation, feedback structs.CommandFeedback) structs.JsonResponse {
	var response = structs.JsonResponse{
		Loaded:     status.Loaded,
		Status:     status.Status,
		ModuleData: maybeGetMetaData(status.MetaChan),
	}
	if status.Loaded {
		execStatus := getExecutionStatus(sim.Execution)
		response.SimulationTime = execStatus.time
		response.RealTimeFactor = execStatus.realTimeFactor
		response.IsRealTimeSimulation = execStatus.isRealTimeSimulation
		response.Module = findModuleData(status, sim.MetaData, sim.Observer)
		response.ConfigDir = status.ConfigDir
		response.TrendSignals = status.TrendSignals
	}
	if (structs.CommandFeedback{}) != feedback {
		response.Feedback = &feedback
	}
	return response
}

func StateUpdateLoop(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *Simulation) {
	for {
		state <- GenerateJsonResponse(simulationStatus, sim, structs.CommandFeedback{})
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, metaData *structs.MetaData, fmuPath string) bool {
	log.Println("Loading: " + fmuPath)
	localSlave := createLocalSlave(fmuPath)
	if localSlave == nil {
		return false
	}
	fmu := metadata.ReadModelDescription(fmuPath)
	index := executionAddSlave(execution, localSlave)
	if index < 0 {
		return false
	}
	fmu.ExecutionIndex = index
	metaData.FMUs = append(metaData.FMUs, fmu)
	return true
}

func addFmuSsd(metaData *structs.MetaData, name string, index int, fmuPath string) {
	//log.Println("Parsing: " + fmuPath)
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
	Execution     *C.cse_execution
	Observer      *C.cse_observer
	TrendObserver *C.cse_observer
	FileObserver  *C.cse_observer
	MetaData      *structs.MetaData
}

func CreateEmptySimulation() Simulation {
	return Simulation{}
}
