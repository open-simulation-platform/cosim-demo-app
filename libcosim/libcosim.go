// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package libcosim

/*
	#cgo CFLAGS: -I${SRCDIR}/../include
	#cgo LDFLAGS: -L${SRCDIR}/../dist/bin -L${SRCDIR}/../dist/lib -lcosimc -lstdc++
	#include <cosim.h>
*/
import "C"
import (
	"cosim-demo-app/structs"
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
	fmt.Printf("Error code %d: %s\n", int(C.cosim_last_error_code()), C.GoString(C.cosim_last_error_message()))
}

func lastErrorMessage() string {
	msg := C.cosim_last_error_message()
	return C.GoString(msg)
}

func lastErrorCode() C.cosim_errc {
	return C.cosim_last_error_code()
}

func createExecution() (execution *C.cosim_execution) {
	startTime := C.cosim_time_point(0.0 * 1e9)
	stepSize := C.cosim_duration(0.1 * 1e9)
	execution = C.cosim_execution_create(startTime, stepSize)
	return execution
}

func createConfigExecution(confDir string) (execution *C.cosim_execution) {
	startTime := C.cosim_time_point(0.0 * 1e9)
	execution = C.cosim_osp_config_execution_create(C.CString(confDir), false, startTime)
	return execution
}

func createSsdExecution(ssdDir string) (execution *C.cosim_execution) {
	startTime := C.cosim_time_point(0.0 * 1e9)
	execution = C.cosim_ssp_execution_create(C.CString(ssdDir), false, startTime)
	return execution
}

type executionStatus struct {
	time                 float64
	realTimeFactor       float64
	realTimeFactorTarget float64
	isRealTimeSimulation bool
	state                string
	lastErrorMessage     string
	lastErrorCode        string
}

func translateState(state C.cosim_execution_state) string {
	switch state {
	case C.COSIM_EXECUTION_STOPPED:
		return "COSIM_EXECUTION_STOPPED"
	case C.COSIM_EXECUTION_RUNNING:
		return "COSIM_EXECUTION_RUNNING"
	case C.COSIM_EXECUTION_ERROR:
		return "COSIM_EXECUTION_ERROR"
	}
	return "UNKNOWN"
}

func translateErrorCode(code C.cosim_errc) string {
	switch code {
	case C.COSIM_ERRC_SUCCESS:
		return "COSIM_ERRC_SUCCESS"
	case C.COSIM_ERRC_UNSPECIFIED:
		return "COSIM_ERRC_UNSPECIFIED"
	case C.COSIM_ERRC_ERRNO:
		return "COSIM_ERRC_ERRNO"
	case C.COSIM_ERRC_INVALID_ARGUMENT:
		return "COSIM_ERRC_INVALID_ARGUMENT"
	case C.COSIM_ERRC_ILLEGAL_STATE:
		return "COSIM_ERRC_ILLEGAL_STATE"
	case C.COSIM_ERRC_OUT_OF_RANGE:
		return "COSIM_ERRC_OUT_OF_RANGE"
	case C.COSIM_ERRC_STEP_TOO_LONG:
		return "COSIM_ERRC_STEP_TOO_LONG"
	case C.COSIM_ERRC_BAD_FILE:
		return "COSIM_ERRC_BAD_FILE"
	case C.COSIM_ERRC_UNSUPPORTED_FEATURE:
		return "COSIM_ERRC_UNSUPPORTED_FEATURE"
	case C.COSIM_ERRC_DL_LOAD_ERROR:
		return "COSIM_ERRC_DL_LOAD_ERROR"
	case C.COSIM_ERRC_MODEL_ERROR:
		return "COSIM_ERRC_MODEL_ERROR"
	case C.COSIM_ERRC_SIMULATION_ERROR:
		return "COSIM_ERRC_SIMULATION_ERROR"
	case C.COSIM_ERRC_ZIP_ERROR:
		return "COSIM_ERRC_ZIP_ERROR"
	}
	return "UNKNOWN"
}

func getExecutionStatus(execution *C.cosim_execution) (execStatus executionStatus) {
	var status C.cosim_execution_status
	success := int(C.cosim_execution_get_status(execution, &status))
	nanoTime := int64(status.current_time)
	execStatus.time = float64(nanoTime) * 1e-9
	execStatus.realTimeFactor = float64(status.real_time_factor)
	execStatus.realTimeFactorTarget = float64(status.real_time_factor_target)
	execStatus.isRealTimeSimulation = int(status.is_real_time_simulation) > 0
	execStatus.state = translateState(status.state)
	if success < 0 || status.state == C.COSIM_EXECUTION_ERROR {
		execStatus.lastErrorMessage = lastErrorMessage()
		execStatus.lastErrorCode = translateErrorCode(lastErrorCode())
	}
	return
}

func createLocalSlave(fmuPath string, instanceName string) *C.cosim_slave {
	return C.cosim_local_slave_create(C.CString(fmuPath), C.CString(instanceName))
}

func createOverrideManipulator() (manipulator *C.cosim_manipulator) {
	manipulator = C.cosim_override_manipulator_create()
	return
}

func executionAddManipulator(execution *C.cosim_execution, manipulator *C.cosim_manipulator) {
	C.cosim_execution_add_manipulator(execution, manipulator)
}

func executionAddSlave(execution *C.cosim_execution, slave *C.cosim_slave) int {
	slaveIndex := C.cosim_execution_add_slave(execution, slave)
	if slaveIndex < 0 {
		printLastError()
		C.cosim_execution_destroy(execution)
	}
	return int(slaveIndex)
}

func executionStart(execution *C.cosim_execution) (bool, string) {
	success := C.cosim_execution_start(execution)
	if int(success) < 0 {
		return false, strCat("Unable to start simulation: ", lastErrorMessage())
	} else {
		return true, "Simulation is running"
	}
}

func executionDestroy(execution *C.cosim_execution) {
	C.cosim_execution_destroy(execution)
}

func localSlaveDestroy(slave *C.cosim_slave) {
	C.cosim_local_slave_destroy(slave)
}

func manipulatorDestroy(manipulator *C.cosim_manipulator) {
	C.cosim_manipulator_destroy(manipulator)
}

func executionStop(execution *C.cosim_execution) (bool, string) {
	success := C.cosim_execution_stop(execution)
	if int(success) < 0 {
		return false, strCat("Unable to stop simulation: ", lastErrorMessage())
	} else {
		return true, "Simulation is paused"
	}
}

func executionEnableRealTime(execution *C.cosim_execution) (bool, string) {
	success := C.cosim_execution_enable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, strCat("Unable to enable real time: ", lastErrorMessage())
	} else {
		return true, "Real time execution enabled"
	}
}

func executionDisableRealTime(execution *C.cosim_execution) (bool, string) {
	success := C.cosim_execution_disable_real_time_simulation(execution)
	if int(success) < 0 {
		return false, strCat("Unable to disable real time: ", lastErrorMessage())
	} else {
		return true, "Real time execution disabled"
	}
}

func executionSetCustomRealTimeFactor(execution *C.cosim_execution, status *structs.SimulationStatus, realTimeFactor string) (bool, string) {
	val, err := strconv.ParseFloat(realTimeFactor, 64)

	if err != nil {
		log.Println(err)
		return false, err.Error()
	}

	if val <= 0.0 {
		return false, "Real time factor target must be greater than 0.0"
	}

	C.cosim_execution_set_real_time_factor_target(execution, C.double(val))

	return true, "Custom real time factor successfully set"
}

func setReal(manipulator *C.cosim_manipulator, slaveIndex int, valueRef int, value float64) (bool, string) {
	vr := make([]C.cosim_value_reference, 1)
	vr[0] = C.cosim_value_reference(valueRef)
	v := make([]C.double, 1)
	v[0] = C.double(value)
	success := C.cosim_manipulator_slave_set_real(manipulator, C.cosim_slave_index(slaveIndex), &vr[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set real variable value: ", lastErrorMessage())
	} else {
		return true, "Successfully set real variable value"
	}
}

func setInteger(manipulator *C.cosim_manipulator, slaveIndex int, valueRef int, value int) (bool, string) {
	vr := make([]C.cosim_value_reference, 1)
	vr[0] = C.cosim_value_reference(valueRef)
	v := make([]C.int, 1)
	v[0] = C.int(value)
	success := C.cosim_manipulator_slave_set_integer(manipulator, C.cosim_slave_index(slaveIndex), &vr[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set integer variable value: ", lastErrorMessage())
	} else {
		return true, "Successfully set integer variable value"
	}
}

func setBoolean(manipulator *C.cosim_manipulator, slaveIndex int, valueRef int, value bool) (bool, string) {
	vr := make([]C.cosim_value_reference, 1)
	vr[0] = C.cosim_value_reference(valueRef)
	v := make([]C.bool, 1)
	v[0] = C.bool(value)
	success := C.cosim_manipulator_slave_set_boolean(manipulator, C.cosim_slave_index(slaveIndex), &vr[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set boolean variable value: ", lastErrorMessage())
	} else {
		return true, "Successfully set boolean variable value"
	}
}

func setString(manipulator *C.cosim_manipulator, slaveIndex int, valueRef int, value string) (bool, string) {
	vr := make([]C.cosim_value_reference, 1)
	vr[0] = C.cosim_value_reference(valueRef)
	v := make([]*C.char, 1)
	v[0] = C.CString(value)
	success := C.cosim_manipulator_slave_set_string(manipulator, C.cosim_slave_index(slaveIndex), &vr[0], C.size_t(1), &v[0])
	if int(success) < 0 {
		return false, strCat("Unable to set boolean variable value", lastErrorMessage())
	} else {
		return true, "Successfully set boolean variable value"
	}
}

func setVariableValue(sim *Simulation, slaveIndex string, valueType string, valueReference string, value string) (bool, string) {
	index, err := strconv.Atoi(slaveIndex)
	if err != nil {
		return false, strCat("Can't parse slave index as integer: ", slaveIndex)
	}
	valueRef, err := strconv.Atoi(valueReference)
	if err != nil {
		return false, strCat("Can't parse value reference as integer: ", valueReference)
	}
	switch valueType {
	case "Real":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println(err)
			return false, strCat("Can't parse value as double: ", value, ", error: ", err.Error())
		} else {
			return setReal(sim.OverrideManipulator, index, valueRef, val)
		}
	case "Integer":
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Println(err)
			return false, strCat("Can't parse value as integer: ", value, ", error: ", err.Error())
		} else {
			return setInteger(sim.OverrideManipulator, index, valueRef, val)
		}
	case "Boolean":
		var val = false
		if "true" == value {
			val = true
		}
		return setBoolean(sim.OverrideManipulator, index, valueRef, val)
	case "String":
		return setString(sim.OverrideManipulator, index, valueRef, value)

	default:
		message := strCat("Can't set this value: ", value)
		fmt.Println(message)
		return false, message
	}
}

func resetVariable(manipulator *C.cosim_manipulator, slaveIndex int, variableType C.cosim_variable_type, valueRef int) (bool, string) {
	vr := make([]C.cosim_value_reference, 1)
	vr[0] = C.cosim_value_reference(valueRef)
	success := C.cosim_manipulator_slave_reset(manipulator, C.cosim_slave_index(slaveIndex), variableType, &vr[0], C.size_t(1))
	if int(success) < 0 {
		return false, strCat("Unable to reset variable value: ", lastErrorMessage())
	} else {
		return true, "Successfully reset variable value"
	}
}

func resetVariableValue(sim *Simulation, slaveIndex string, valueType string, valueReference string) (bool, string) {
	index, err := strconv.Atoi(slaveIndex)
	if err != nil {
		return false, strCat("Can't parse slave index as integer: ", slaveIndex)
	}
	valueRef, err := strconv.Atoi(valueReference)
	if err != nil {
		return false, strCat("Can't parse value reference as integer: ", valueReference)
	}
	switch valueType {
	case "Real":
		return resetVariable(sim.OverrideManipulator, index, C.COSIM_VARIABLE_TYPE_REAL, valueRef)
	case "Integer":
		return resetVariable(sim.OverrideManipulator, index, C.COSIM_VARIABLE_TYPE_INTEGER, valueRef)
	case "Boolean":
		return resetVariable(sim.OverrideManipulator, index, C.COSIM_VARIABLE_TYPE_BOOLEAN, valueRef)
	case "String":
		return resetVariable(sim.OverrideManipulator, index, C.COSIM_VARIABLE_TYPE_STRING, valueRef)
	default:
		message := strCat("Can't reset variable with type ", valueType, " and value reference ", valueReference, " for slave with index ", slaveIndex)
		log.Println(message)
		return false, message
	}
}

func parseCausality(causality C.cosim_variable_causality) (string, error) {
	switch causality {
	case C.COSIM_VARIABLE_CAUSALITY_INPUT:
		return "input", nil
	case C.COSIM_VARIABLE_CAUSALITY_OUTPUT:
		return "output", nil
	case C.COSIM_VARIABLE_CAUSALITY_PARAMETER:
		return "parameter", nil
	case C.COSIM_VARIABLE_CAUSALITY_CALCULATEDPARAMETER:
		return "calculatedParameter", nil
	case C.COSIM_VARIABLE_CAUSALITY_LOCAL:
		return "local", nil
	case C.COSIM_VARIABLE_CAUSALITY_INDEPENDENT:
		return "independent", nil
	}
	return "", errors.New("Unable to parse variable causality")
}

func parseVariability(variability C.cosim_variable_variability) (string, error) {
	switch variability {
	case C.COSIM_VARIABLE_VARIABILITY_CONSTANT:
		return "constant", nil
	case C.COSIM_VARIABLE_VARIABILITY_FIXED:
		return "fixed", nil
	case C.COSIM_VARIABLE_VARIABILITY_TUNABLE:
		return "tunable", nil
	case C.COSIM_VARIABLE_VARIABILITY_DISCRETE:
		return "discrete", nil
	case C.COSIM_VARIABLE_VARIABILITY_CONTINUOUS:
		return "continuous", nil
	}
	return "", errors.New("Unable to parse variable variability")
}

func parseType(valueType C.cosim_variable_type) (string, error) {
	switch valueType {
	case C.COSIM_VARIABLE_TYPE_REAL:
		return "Real", nil
	case C.COSIM_VARIABLE_TYPE_INTEGER:
		return "Integer", nil
	case C.COSIM_VARIABLE_TYPE_STRING:
		return "String", nil
	case C.COSIM_VARIABLE_TYPE_BOOLEAN:
		return "Boolean", nil
	}
	return "", errors.New("unable to parse variable type")
}

func addVariableMetadata(execution *C.cosim_execution, fmu *structs.FMU) error {
	nVariables := C.cosim_slave_get_num_variables(execution, C.cosim_slave_index(fmu.ExecutionIndex))
	if int(nVariables) < 0 {
		return errors.New("invalid slave index to find variables for")
	} else if int(nVariables) == 0 {
		log.Println("No variables for ", fmu.Name)
		return nil
	}

	var variables = make([]C.cosim_variable_description, int(nVariables))
	nVariablesRead := C.cosim_slave_get_variables(execution, C.cosim_slave_index(fmu.ExecutionIndex), &variables[0], C.size_t(nVariables))
	if int(nVariablesRead) < 0 {
		return errors.New(strCat("Unable to get variables for slave with name ", fmu.Name))
	}
	for _, variable := range variables[0:int(nVariablesRead)] {
		name := C.GoString(&variable.name[0])
		ref := int(variable.reference)
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
			ref,
			causality,
			variability,
			valueType,
		})
	}
	return nil
}

func fetchManipulatedVariables(execution *C.cosim_execution) []structs.ManipulatedVariable {
	nVars := int(C.cosim_get_num_modified_variables(execution))
	if nVars <= 0 {
		return nil
	}

	var variables = make([]C.cosim_variable_id, nVars)
	numVars := int(C.cosim_get_modified_variables(execution, &variables[0], C.size_t(nVars)))

	if numVars < 0 {
		log.Println("Error while fetching modified variables: ", lastErrorMessage())
		return nil
	}

	var varStructs = make([]structs.ManipulatedVariable, numVars)
	for n, variable := range variables[0:numVars] {
		slaveIndex := int(variable.slave_index)
		valueReference := int(variable.value_reference)
		variableType, err := parseType(variable._type)

		if err != nil {
			log.Println("Problem parsing variable type: ", variable._type)
			return nil
		}

		varStructs[n] = structs.ManipulatedVariable{
			slaveIndex,
			variableType,
			valueReference,
		}
	}

	return varStructs
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
	sim.LocalSlaves = []*C.cosim_slave{}
	sim.Observer = nil
	sim.TrendObserver = nil
	sim.FileObserver = nil
	sim.OverrideManipulator = nil
	sim.ScenarioManager = nil
	sim.MetaData = &structs.MetaData{}
	return true, "Simulation teardown successful"
}

func validateConfigPath(configPath string) (bool, string) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, strCat(configPath, " does not exist")
	}
	return true, ""
}

func isDirectory(configPath string) bool {
	fi, _ := os.Stat(configPath)
	return fi.Mode().IsDir()
}

func validateConfigDir(fmuDir string) (bool, string) {
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
	var hasSSD = false
	for _, f := range files {
		if f.Name() == "SystemStructure.ssd" {
			hasSSD = true
		}
	}
	var hasOspXml = false
	for _, f := range files {
		if f.Name() == "OspSystemStructure.xml" {
			hasOspXml = true
		}
	}
	if !hasFMUs && !hasSSD && !hasOspXml {
		return false, strCat(fmuDir, " does not contain any FMUs, OspSystemStructure.xml nor a SystemStructure.ssd file")
	}
	return true, ""
}

type configuration struct {
	configFile  string
	configDir   string
	isSsdConfig bool
	isOspConfig bool
}

func checkConfiguration(configPath string) (valid bool, message string, config configuration) {
	if valid, message := validateConfigPath(configPath); !valid {
		return false, message, config
	}

	if isDirectory(configPath) {
		if valid, message := validateConfigDir(configPath); !valid {
			return false, message, config
		}
		config.configDir = configPath
		if hasFile(configPath, "OspSystemStructure.xml") {
			config.configFile = filepath.Join(configPath, "OspSystemStructure.xml")
			config.isOspConfig = true
		} else if hasFile(configPath, "SystemStructure.ssd") {
			config.configFile = filepath.Join(configPath, "SystemStructure.ssd")
			config.isSsdConfig = true
		}
	} else {
		config.configFile = configPath
		config.configDir = filepath.Dir(configPath)
		if strings.HasSuffix(configPath, ".xml") {
			config.isOspConfig = true
		} else if strings.HasSuffix(configPath, ".ssd") {
			config.isSsdConfig = true
		} else {
			return false, "Given file path does not have a recognized format (xml, ssd): " + configPath, config
		}
	}

	return true, "", config
}

func initializeSimulation(sim *Simulation, status *structs.SimulationStatus, configPath string, logDir string) (bool, string, string) {
	valid, message, config := checkConfiguration(configPath)
	if !valid {
		return false, message, ""
	}

	var execution *C.cosim_execution
	if config.isOspConfig {
		execution = createConfigExecution(config.configFile)
		if execution == nil {
			return false, strCat("Could not create execution from OspSystemStructure.xml file: ", lastErrorMessage()), ""
		}
	} else if config.isSsdConfig {
		execution = createSsdExecution(config.configFile)
		if execution == nil {
			return false, strCat("Could not create execution from SystemStructure.ssd file: ", lastErrorMessage()), ""
		}
	} else {
		execution = createExecution()
		if execution == nil {
			return false, strCat("Could not create execution: ", lastErrorMessage()), ""
		}
		paths := getFmuPaths(config.configDir)
		for _, path := range paths {
			slave, err := addFmu(execution, path)
			if err != nil {
				return false, strCat("Could not add FMU to execution: ", err.Error()), ""
			} else {
				sim.LocalSlaves = append(sim.LocalSlaves, slave)
			}
		}
	}

	metaData := structs.MetaData{
		FMUs: []structs.FMU{},
	}
	err := addMetadata(execution, &metaData)
	if err != nil {
		return false, err.Error(), ""
	}

	observer := createObserver()
	executionAddObserver(execution, observer)

	trendObserver := createTrendObserver()
	executionAddObserver(execution, trendObserver)

	var fileObserver *C.cosim_observer
	if len(logDir) > 0 {
		if hasFile(config.configDir, "LogConfig.xml") {
			fileObserver = createFileObserverFromCfg(logDir, filepath.Join(config.configDir, "LogConfig.xml"))
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

	setupPlotsFromConfig(sim, status, config.configDir)

	return true, "Simulation loaded successfully", config.configDir
}

func resetSimulation(sim *Simulation, status *structs.SimulationStatus, configPath string, logDir string) (bool, string, string) {
    var success = false
    var message = ""
    var configDir = ""

    success, message = executionStop(sim.Execution)
    log.Println(message)

    if success{
        status.Loaded = false
        status.Status = "stopped"
        status.ConfigDir = ""
        status.Trends = []structs.Trend{}
        status.Module = ""
        success, message = simulationTeardown(sim)
        log.Println(message)
    }

    success, message, configDir = initializeSimulation(sim, status, configPath, logDir)
    log.Println(message)

    return success, message, configDir
}

func executeCommand(cmd []string, sim *Simulation, status *structs.SimulationStatus) (shorty structs.ShortLivedData, feedback structs.CommandFeedback) {
	var success = false
	var message = "No feedback implemented for this command"
	switch cmd[0] {
	case "load":
		status.Loading = true
		var configDir string
		success, message, configDir = initializeSimulation(sim, status, cmd[1], cmd[2])
		if success {
			status.Loaded = true
			status.ConfigDir = configDir
			status.Status = "pause"
			shorty.ModuleData = sim.MetaData
			scenarios := findScenarios(status)
			shorty.Scenarios = &scenarios
		}
		status.Loading = false
	case "teardown":
		status.Loaded = false
		status.Status = "stopped"
		status.ConfigDir = ""
		status.Trends = []structs.Trend{}
		status.Module = ""
		success, message = simulationTeardown(sim)
		shorty.ModuleData = sim.MetaData
	case "reset":
        status.Loading = true
        var configDir string
        success, message, configDir = resetSimulation(sim, status, cmd[1], cmd[2])
        if success {
            status.Loaded = true
        	status.ConfigDir = configDir
        	status.Status = "pause"
        	shorty.ModuleData = sim.MetaData
        	scenarios := findScenarios(status)
        	shorty.Scenarios = &scenarios
        }
        status.Loading = false
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
	case "set-custom-realtime-factor":
		success, message = executionSetCustomRealTimeFactor(sim.Execution, status, cmd[1])
	case "newtrend":
		success, message = addNewTrend(status, cmd[1], cmd[2])
	case "addtotrend":
		success, message = addToTrend(sim, status, cmd[1], cmd[2], cmd[3])
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
		success, message = setVariableValue(sim, cmd[1], cmd[2], cmd[3], cmd[4])
	case "reset-value":
		success, message = resetVariableValue(sim, cmd[1], cmd[2], cmd[3])
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
		for j := 2; j < (len(cmd) - 3); j += 4 {
			name := cmd[j]
			caus := cmd[j+1]
			typ := cmd[j+2]
			vr := cmd[j+3]
			valRef, err := strconv.Atoi(vr)
			if err != nil {
				message = strCat("Could not parse value reference from: ", vr, ", ", err.Error())
				log.Println(message)
				success = false
			} else {
				variables = append(variables,
					structs.Variable{
						Name:           name,
						Causality:      caus,
						Type:           typ,
						ValueReference: valRef,
					})
			}
		}
	} else {
		message = "Successfully reset signal subscriptions"
	}
	status.SignalSubscriptions = variables
	return success, message
}

func findFmu(metaData *structs.MetaData, moduleName string) (foundFmu structs.FMU, err error) {
	for _, fmu := range metaData.FMUs {
		if fmu.Name == moduleName {
			foundFmu = fmu
			return foundFmu, nil
		}
	}
	return foundFmu, errors.New("Simulator with name " + moduleName + " does not exist.")
}

func findVariable(fmu structs.FMU, variableName string) (foundVariable structs.Variable, err error) {
	for _, variable := range fmu.Variables {
		if variable.Name == variableName {
			foundVariable = variable
			return foundVariable, nil
		}
	}
	return foundVariable, errors.New("Variable with name " + variableName + " does not exist for simulator " + fmu.Name)
}

func findModuleData(status *structs.SimulationStatus, metaData *structs.MetaData, observer *C.cosim_observer) (module structs.Module) {
	if len(status.SignalSubscriptions) > 0 {

		slave, err := findFmu(metaData, status.Module)
		if err != nil {
			log.Println(err.Error())
			return
		}
		slaveIndex := slave.ExecutionIndex
		realSignals := observerGetReals(observer, status.SignalSubscriptions, slaveIndex)
		intSignals := observerGetIntegers(observer, status.SignalSubscriptions, slaveIndex)
		boolSignals := observerGetBooleans(observer, status.SignalSubscriptions, slaveIndex)
		stringSignals := observerGetStrings(observer, status.SignalSubscriptions, slaveIndex)
		var signals []structs.Signal
		signals = append(signals, realSignals...)
		signals = append(signals, intSignals...)
		signals = append(signals, boolSignals...)
		signals = append(signals, stringSignals...)

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
		Loading: status.Loading,
		Loaded:  status.Loaded,
		Status:  status.Status,
	}
	if status.Loaded {
		execStatus := getExecutionStatus(sim.Execution)
		response.ExecutionState = execStatus.state
		response.LastErrorCode = execStatus.lastErrorCode
		response.LastErrorMessage = execStatus.lastErrorMessage
		response.SimulationTime = execStatus.time
		response.RealTimeFactor = execStatus.realTimeFactor
		response.RealTimeFactorTarget = execStatus.realTimeFactorTarget
		response.IsRealTimeSimulation = execStatus.isRealTimeSimulation
		response.Module = findModuleData(status, sim.MetaData, sim.Observer)
		response.ConfigDir = status.ConfigDir
		response.Trends = status.Trends
		response.ManipulatedVariables = fetchManipulatedVariables(sim.Execution)
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

func addFmu(execution *C.cosim_execution, fmuPath string) (*C.cosim_slave, error) {
	baseName := filepath.Base(fmuPath)
	instanceName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	fmt.Printf("Creating instance %s from %s\n", instanceName, fmuPath)
	localSlave := createLocalSlave(fmuPath, instanceName)
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

func addMetadata(execution *C.cosim_execution, metaData *structs.MetaData) error {
	nSlaves := C.cosim_execution_get_num_slaves(execution)
	var slaveInfos = make([]C.cosim_slave_info, int(nSlaves))
	C.cosim_execution_get_slave_infos(execution, &slaveInfos[0], nSlaves)
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
	Execution           *C.cosim_execution
	Observer            *C.cosim_observer
	TrendObserver       *C.cosim_observer
	FileObserver        *C.cosim_observer
	OverrideManipulator *C.cosim_manipulator
	ScenarioManager     *C.cosim_manipulator
	MetaData            *structs.MetaData
	LocalSlaves         []*C.cosim_slave
}

func CreateEmptySimulation() Simulation {
	return Simulation{}
}

func SetupLogging() {
	success := C.cosim_log_setup_simple_console_logging()
	if int(success) < 0 {
		log.Println("Could not set up console logging!")
	} else {
		C.cosim_log_set_output_level(C.COSIM_LOG_SEVERITY_INFO)
		log.Println("Console logging set up with severity: INFO")
	}
}
