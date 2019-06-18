package cse

/*
	#cgo CFLAGS: -I${SRCDIR}/../include
	#cgo LDFLAGS: -L${SRCDIR}/../dist/bin -L${SRCDIR}/../dist/lib -lcsecorec -lstdc++
	#include <cse.h>
*/
import "C"
import (
	"cse-server-go/structs"
	"errors"
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

func lastErrorMessage() string {
	msg := C.cse_last_error_message()
	return C.GoString(msg)
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

func createOverrideManipulator() (manipulator *C.cse_manipulator) {
	manipulator = C.cse_override_manipulator_create()
	return
}

func executionAddManipulator(execution *C.cse_execution, manipulator *C.cse_manipulator) {
	C.cse_execution_add_manipulator(execution, manipulator)
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
		return false, strCat("Unable to start simulation: " + lastErrorMessage())
	} else {
		return true, "Simulation is running"
	}
}

func executionDestroy(execution *C.cse_execution) {
	C.cse_execution_destroy(execution)
}

func localSlaveDestroy(slave *C.cse_slave) {
	C.cse_local_slave_destroy(slave)
}

func manipulatorDestroy(manipulator *C.cse_manipulator) {
	C.cse_manipulator_destroy(manipulator)
}

func executionStop(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_stop(execution)
	if int(success) < 0 {
		return false, strCat("Unable to stop simulation: " + lastErrorMessage())
	} else {
		return true, "Simulation is paused"
	}
}

func executionEnableRealTime(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_enable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, strCat("Unable to enable real time: " + lastErrorMessage())
	} else {
		return true, "Real time execution enabled"
	}
}

func executionDisableRealTime(execution *C.cse_execution) (bool, string) {
	success := C.cse_execution_disable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, strCat("Unable to disable real time: " + lastErrorMessage())
	} else {
		return true, "Real time execution disabled"
	}
}

func setReal(manipulator *C.cse_manipulator, slaveIndex int, variableIndex int, value float64) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.double, 1)
	v[0] = C.double(value)
	success := C.cse_manipulator_slave_set_real(manipulator, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set real variable value: " + lastErrorMessage())
	} else {
		return true, "Successfully set real variable value"
	}
}

func setInteger(manipulator *C.cse_manipulator, slaveIndex int, variableIndex int, value int) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	v := make([]C.int, 1)
	v[0] = C.int(value)
	success := C.cse_manipulator_slave_set_integer(manipulator, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set integer variable value: " + lastErrorMessage())
	} else {
		return true, "Successfully set integer variable value"
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
			return setReal(sim.OverrideManipulator, fmu.ExecutionIndex, varIndex, val)
		}
	case "Integer":
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Println(err)
			return false, err.Error()
		} else {
			return setInteger(sim.OverrideManipulator, fmu.ExecutionIndex, varIndex, val)
		}
	default:
		message := strCat("Can't set this value: ", value)
		fmt.Println(message)
		return false, message
	}
}

func resetReal(manipulator *C.cse_manipulator, slaveIndex int, variableIndex int) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	success := C.cse_manipulator_slave_reset_real(manipulator, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1))
	if int(success) < 0 {
		return false, strCat("Unable to reset real variable value: " + lastErrorMessage())
	} else {
		return true, "Successfully reset real variable value"
	}
}

func resetInteger(manipulator *C.cse_manipulator, slaveIndex int, variableIndex int) (bool, string) {
	vi := make([]C.cse_variable_index, 1)
	vi[0] = C.cse_variable_index(variableIndex)
	success := C.cse_manipulator_slave_reset_integer(manipulator, C.cse_slave_index(slaveIndex), &vi[0], C.size_t(1))
	if int(success) < 0 {
		return false, strCat("Unable to reset integer variable value: " + lastErrorMessage())
	} else {
		return true, "Successfully reset integer variable value"
	}
}

func resetVariableValue(sim *Simulation, module string, signal string, causality string, valueType string) (bool, string) {
	fmu := findFmu(sim.MetaData, module)
	varIndex := findVariableIndex(fmu, signal, causality, valueType)
	switch valueType {
	case "Real":
		return resetReal(sim.OverrideManipulator, fmu.ExecutionIndex, varIndex)
	case "Integer":
		return resetInteger(sim.OverrideManipulator, fmu.ExecutionIndex, varIndex)
	default:
		message := strCat("Can't reset this variable: ", module, " - ", signal)
		fmt.Println(message)
		return false, message
	}
}

func parseCausality(causality C.cse_variable_causality) (string, error) {
	switch causality {
	case C.CSE_VARIABLE_CAUSALITY_INPUT:
		return "input", nil
	case C.CSE_VARIABLE_CAUSALITY_OUTPUT:
		return "output", nil
	case C.CSE_VARIABLE_CAUSALITY_PARAMETER:
		return "parameter", nil
	case C.CSE_VARIABLE_CAUSALITY_CALCULATEDPARAMETER:
		return "calculatedParameter", nil
	case C.CSE_VARIABLE_CAUSALITY_LOCAL:
		return "local", nil
	case C.CSE_VARIABLE_CAUSALITY_INDEPENDENT:
		return "independent", nil
	}
	return "", errors.New("Unable to parse variable causality")
}

func parseVariability(variability C.cse_variable_variability) (string, error) {
	switch variability {
	case C.CSE_VARIABLE_VARIABILITY_CONSTANT:
		return "constant", nil
	case C.CSE_VARIABLE_VARIABILITY_FIXED:
		return "fixed", nil
	case C.CSE_VARIABLE_VARIABILITY_TUNABLE:
		return "tunable", nil
	case C.CSE_VARIABLE_VARIABILITY_DISCRETE:
		return "discrete", nil
	case C.CSE_VARIABLE_VARIABILITY_CONTINUOUS:
		return "continuous", nil
	}
	return "", errors.New("Unable to parse variable variability")
}

func parseType(valueType C.cse_variable_type) (string, error) {
	switch valueType {
	case C.CSE_VARIABLE_TYPE_REAL:
		return "Real", nil
	case C.CSE_VARIABLE_TYPE_INTEGER:
		return "Integer", nil
	case C.CSE_VARIABLE_TYPE_STRING:
		return "String", nil
	case C.CSE_VARIABLE_TYPE_BOOLEAN:
		return "Boolean", nil
	}
	return "", errors.New("unable to parse variable type")
}

func addVariableMetadata(execution *C.cse_execution, fmu *structs.FMU) error {
	nVariables := C.cse_slave_get_num_variables(execution, C.cse_slave_index(fmu.ExecutionIndex))
	if int(nVariables) < 0 {
		return errors.New("invalid slave index to find variables for")
	} else if int(nVariables) == 0 {
		log.Println("No variables for ", fmu.Name)
		return nil
	}

	var variables = make([]C.cse_variable_description, int(nVariables))
	nVariablesRead := C.cse_slave_get_variables(execution, C.cse_slave_index(fmu.ExecutionIndex), &variables[0], C.size_t(nVariables))
	if int(nVariablesRead) < 0 {
		return errors.New(strCat("Unable to get variables for slave with name ", fmu.Name))
	}
	for _, variable := range variables[0:int(nVariablesRead)] {
		name := C.GoString(&variable.name[0])
		index := int(variable.index)
		causality, err := parseCausality(variable.causality)
		if err != nil {
			return errors.New(strCat("Problem parsing causality for slave ", fmu.Name, ", variable ", name))
		}
		variability, err := parseVariability(variable.variability)
		if err != nil {
			return errors.New(strCat("Problem parsing variability for slave ", fmu.Name, ", variable ", name))
		}
		valueType, err := parseType(variable._type)
		if err != nil {
			return errors.New(strCat("Problem parsing type for slave ", fmu.Name, ", variable ", name))
		}
		fmu.Variables = append(fmu.Variables, structs.Variable{
			name,
			index,
			causality,
			variability,
			valueType,
		})
	}
	return nil
}

func simulationTeardown(sim *Simulation) (bool, string) {
	executionDestroy(sim.Execution)
	observerDestroy(sim.Observer)
	observerDestroy(sim.TrendObserver)
	if nil != sim.FileObserver {
		observerDestroy(sim.FileObserver)
	}
	manipulatorDestroy(sim.OverrideManipulator)
	manipulatorDestroy(sim.ScenarioManager)
	for _, slave := range sim.LocalSlaves {
		localSlaveDestroy(slave)
	}

	sim.Execution = nil
	sim.LocalSlaves = []*C.cse_slave{}
	sim.Observer = nil
	sim.TrendObserver = nil
	sim.FileObserver = nil
	sim.OverrideManipulator = nil
	sim.ScenarioManager = nil
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
	if hasFile(fmuDir, "SystemStructure.ssd") {
		execution = createSsdExecution(fmuDir)
		if execution == nil {
			return false, strCat("Could not create execution from SystemStructure.ssd file: ", lastErrorMessage())
		}
	} else {
		execution = createExecution()
		if execution == nil {
			return false, strCat("Could not create execution: ", lastErrorMessage())
		}
		paths := getFmuPaths(fmuDir)
		for _, path := range paths {
			slave, err := addFmu(execution, path)
			if err != nil {
				return false, strCat("Could not add FMU to execution: ", err.Error())
			} else {
				sim.LocalSlaves = append(sim.LocalSlaves, slave)
			}
		}
	}

	err := addMetadata(execution, &metaData)
	if err != nil {
		return false, err.Error()
	}

	observer := createObserver()
	executionAddObserver(execution, observer)

	trendObserver := createTrendObserver()
	executionAddObserver(execution, trendObserver)

	var fileObserver *C.cse_observer
	if len(logDir) > 0 {
		if hasFile(fmuDir, "LogConfig.xml") {
			fileObserver = createFileObserverFromCfg(logDir, filepath.Join(fmuDir, "LogConfig.xml"))
		} else {
			fileObserver = createFileObserver(logDir)
		}
		executionAddObserver(execution, fileObserver)
	}

	manipulator := createOverrideManipulator()
	executionAddManipulator(execution, manipulator)

	scenarioManager := createScenarioManager()
	executionAddManipulator(execution, scenarioManager)

	sim.Execution = execution
	sim.Observer = observer
	sim.TrendObserver = trendObserver
	sim.FileObserver = fileObserver
	sim.OverrideManipulator = manipulator
	sim.ScenarioManager = scenarioManager
	sim.MetaData = &metaData
	return true, "Simulation loaded successfully"
}

func executeCommand(cmd []string, sim *Simulation, status *structs.SimulationStatus) (shorty structs.ShortLivedData, feedback structs.CommandFeedback) {
	var success = false
	var message = "No feedback implemented for this command"
	switch cmd[0] {
	case "load":
		success, message = initializeSimulation(sim, cmd[1], cmd[2])
		if success {
			status.Loaded = true
			status.ConfigDir = cmd[1]
			status.Status = "pause"
			shorty.ModuleData = sim.MetaData
			scenarios := findScenarios(status)
			shorty.Scenarios = &scenarios
		}
	case "teardown":
		status.Loaded = false
		status.Status = "stopped"
		status.ConfigDir = ""
		status.Trends = []structs.Trend{}
		status.Module = ""
		success, message = simulationTeardown(sim)
		shorty.ModuleData = sim.MetaData
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
	case "newtrend":
		success, message = addNewTrend(status, cmd[1], cmd[2])
	case "addtotrend":
		success, message = addToTrend(sim, status, cmd[1], cmd[2], cmd[3], cmd[4], cmd[5], cmd[6])
	case "untrend":
		success, message = removeAllFromTrend(sim, status, cmd[1])
	case "removetrend":
		success, message = removeTrend(status, cmd[1])
	case "active-trend":
		success, message = activeTrend(status, cmd[1])
	case "setlabel":
		success, message = setTrendLabel(status, cmd[1], cmd[2])
	case "trend-zoom":
		idx, _ := strconv.Atoi(cmd[1])
		status.Trends[idx].Spec = structs.TrendSpec{Auto: false, Begin: parseFloat(cmd[2]), End: parseFloat(cmd[3])}
		success = true
		message = strCat("Plotting values from ", cmd[2], " to ", cmd[3])
	case "trend-zoom-reset":
		idx, _ := strconv.Atoi(cmd[1])
		status.Trends[idx].Spec = structs.TrendSpec{Auto: true, Range: parseFloat(cmd[2])}
		success = true
		message = strCat("Plotting last ", cmd[2], " seconds")
	case "set-value":
		success, message = setVariableValue(sim, cmd[1], cmd[2], cmd[3], cmd[4], cmd[5])
	case "reset-value":
		success, message = resetVariableValue(sim, cmd[1], cmd[2], cmd[3], cmd[4])
	case "get-module-data":
		shorty.ModuleData = sim.MetaData
		scenarios := findScenarios(status)
		shorty.Scenarios = &scenarios
		success = true
		message = "Fetched metadata"
	case "signals":
		success, message = setSignalSubscriptions(status, cmd)
	case "load-scenario":
		success, message = loadScenario(sim, status, cmd[1])
	case "abort-scenario":
		success, message = abortScenario(sim.ScenarioManager)
	case "parse-scenario":
		scenario, err := parseScenario(status, cmd[1])
		if err != nil {
			success = false
			message = err.Error()
		} else {
			success = true
			message = "Successfully parsed scenario"
			shorty.Scenario = &scenario
		}
	default:
		message = "Unknown command, this is not good"
		fmt.Println(message, cmd)
	}
	return shorty, structs.CommandFeedback{Success: success, Message: message, Command: cmd[0]}
}

func CommandLoop(state chan structs.JsonResponse, sim *Simulation, command chan []string, status *structs.SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			shorty, feedback := executeCommand(cmd, sim, status)
			state <- GenerateJsonResponse(status, sim, feedback, shorty)
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

func GenerateJsonResponse(status *structs.SimulationStatus, sim *Simulation, feedback structs.CommandFeedback, shorty structs.ShortLivedData) structs.JsonResponse {
	var response = structs.JsonResponse{
		Loaded: status.Loaded,
		Status: status.Status,
	}
	if status.Loaded {
		execStatus := getExecutionStatus(sim.Execution)
		response.SimulationTime = execStatus.time
		response.RealTimeFactor = execStatus.realTimeFactor
		response.IsRealTimeSimulation = execStatus.isRealTimeSimulation
		response.Module = findModuleData(status, sim.MetaData, sim.Observer)
		response.ConfigDir = status.ConfigDir
		response.Trends = status.Trends
		if sim.ScenarioManager != nil && isScenarioRunning(sim.ScenarioManager) {
			response.RunningScenario = status.CurrentScenario
		}

	}
	if (structs.CommandFeedback{}) != feedback {
		response.Feedback = &feedback
	}
	if (structs.ShortLivedData{} != shorty) {
		if shorty.Scenarios != nil {
			response.Scenarios = shorty.Scenarios
		}
		if shorty.Scenario != nil {
			response.Scenario = shorty.Scenario
		}
		if shorty.ModuleData != nil {
			response.ModuleData = shorty.ModuleData
		}
	}
	return response
}

func StateUpdateLoop(state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *Simulation) {
	for {
		state <- GenerateJsonResponse(simulationStatus, sim, structs.CommandFeedback{}, structs.ShortLivedData{})
		time.Sleep(1000 * time.Millisecond)
	}
}

func addFmu(execution *C.cse_execution, fmuPath string) (*C.cse_slave, error) {
	log.Println("Loading: " + fmuPath)
	localSlave := createLocalSlave(fmuPath)
	if localSlave == nil {
		printLastError()
		return nil, errors.New(strCat("Unable to create slave from fmu: ", fmuPath))
	}
	index := executionAddSlave(execution, localSlave)
	if index < 0 {
		return nil, errors.New(strCat("Unable to add slave to execution: ", fmuPath))
	}
	return localSlave, nil
}

func addMetadata(execution *C.cse_execution, metaData *structs.MetaData) error {
	nSlaves := C.cse_execution_get_num_slaves(execution)
	var slaveInfos = make([]C.cse_slave_info, int(nSlaves))
	C.cse_execution_get_slave_infos(execution, &slaveInfos[0], nSlaves)
	for _, info := range slaveInfos {
		name := C.GoString(&info.name[0])
		index := int(info.index)
		fmu := structs.FMU{}
		fmu.Name = name
		fmu.ExecutionIndex = index
		err := addVariableMetadata(execution, &fmu)
		metaData.FMUs = append(metaData.FMUs, fmu)
		if err != nil {
			return err
		}
	}
	return nil
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

func hasFile(folder string, fileName string) bool {
	info, e := os.Stat(folder)
	if os.IsNotExist(e) {
		fmt.Println("Folder does not exist: ", folder)
		return false
	} else if !info.IsDir() {
		fmt.Println("Folder is not a directory: ", folder)
		return false
	} else {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if f.Name() == fileName {
				return true
			}
		}
	}
	fmt.Println("Folder does not contain file: ", fileName, folder)
	return false
}

type Simulation struct {
	Execution           *C.cse_execution
	Observer            *C.cse_observer
	TrendObserver       *C.cse_observer
	FileObserver        *C.cse_observer
	OverrideManipulator *C.cse_manipulator
	ScenarioManager     *C.cse_manipulator
	MetaData            *structs.MetaData
	LocalSlaves         []*C.cse_slave
}

func CreateEmptySimulation() Simulation {
	return Simulation{}
}
