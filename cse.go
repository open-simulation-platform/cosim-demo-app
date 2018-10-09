package main

/*
	#include <cse.h>
*/
import "C"
import (
	"fmt"
	"time"
)

// UGLY GLOBAL VARIABLE
var lastOutValue = 0.0
var lastSamplesValue = []C.double{}

func printLastError() {
	fmt.Printf("Error code %d: %s\n", int(C.cse_last_error_code()), C.GoString(C.cse_last_error_message()))
}

func createExecution() (execution *C.cse_execution) {
	execution = C.cse_execution_create(0.0, 0.01)
	return execution
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

func observerAddSlave(observer *C.cse_observer, slave *C.cse_slave) int {
	slaveIndex := C.cse_observer_add_slave(observer, slave)
	if slaveIndex < 0 {
		printLastError()
		//C.cse_observer_destroy(observer)
	}
	return int(slaveIndex)
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

func observerGetReals(observer *C.cse_observer, fmu FMU) (realSignals []Signal) {
	var realValueRefs []C.uint
	var realVariables []Variable
	var numReals int
	for i := range fmu.Variables {
		if fmu.Variables[i].Type == "Real" {
			ref := C.uint(fmu.Variables[i].ValueReference)
			realValueRefs = append(realValueRefs, ref)
			realVariables = append(realVariables, fmu.Variables[i])
			numReals++
		}
	}

	if numReals > 0 {
		realOutVal := make([]C.double, numReals)
		C.cse_observer_slave_get_real(observer, C.int(fmu.ObserverIndex), &realValueRefs[0], C.ulonglong(numReals), &realOutVal[0])

		realSignals = make([]Signal, numReals)
		for k := range realVariables {
			realSignals[k] = Signal{
				Name:      realVariables[k].Name,
				Causality: realVariables[k].Causality,
				Type:      realVariables[k].Type,
				Value:     float64(realOutVal[k]),
			}
		}
	}
	return realSignals
}

func observerGetIntegers(observer *C.cse_observer, fmu FMU) (intSignals []Signal) {
	var intValueRefs []C.uint
	var intVariables []Variable
	var numIntegers int
	for i := range fmu.Variables {
		if fmu.Variables[i].Type == "Integer" {
			ref := C.uint(fmu.Variables[i].ValueReference)
			intValueRefs = append(intValueRefs, ref)
			intVariables = append(intVariables, fmu.Variables[i])
			numIntegers++
		}
	}

	if numIntegers > 0 {
		intOutVal := make([]C.int, numIntegers)
		C.cse_observer_slave_get_integer(observer, C.int(fmu.ObserverIndex), &intValueRefs[0], C.ulonglong(numIntegers), &intOutVal[0])

		intSignals = make([]Signal, numIntegers)
		for k := range intVariables {
			intSignals[k] = Signal{
				Name:      intVariables[k].Name,
				Causality: intVariables[k].Causality,
				Type:      intVariables[k].Type,
				Value:     int(intOutVal[k]),
			}
		}
	}
	return intSignals
}

func observerGetRealSamples(observer *C.cse_observer, nSamples int, signal *TrendSignal) {
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

func polling(observer *C.cse_observer, status *SimulationStatus) {
	for {
		if len(status.TrendSignals) > 0 {
			observerGetRealSamples(observer, 10, &status.TrendSignals[0])
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func simulate(execution *C.cse_execution, command chan []string, status *SimulationStatus) {
	for {
		select {
		case cmd := <-command:
			switch cmd[0] {
			case "stop":
				return
			case "pause":
				executionStop(execution)
				status.Status = "pause"
			case "play":
				executionStart(execution)
				status.Status = "play"
			case "trend":
				status.TrendSignals = append(status.TrendSignals, TrendSignal{cmd[1], cmd[2], nil, nil})
			case "untrend":
				status.TrendSignals = []TrendSignal{}
			case "module":
				status.Module = Module{
					Name: cmd[1],
				}
			default:
				fmt.Println("Empty command, mildt sagt not good: ", cmd)
			}
		}
	}
}
