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

func printLastError() () {
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

func observerAddSlave(observer *C.cse_observer, slave *C.cse_slave) (slaveIndex C.int) {
	slaveIndex = C.cse_observer_add_slave(observer, slave)
	if slaveIndex < 0 {
		printLastError()
		//C.cse_observer_destroy(observer)
	}
	return
}

func executionAddSlave(execution *C.cse_execution, slave *C.cse_slave) (slaveIndex C.int) {
	slaveIndex = C.cse_execution_add_slave(execution, slave)
	if slaveIndex < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	return
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
func observerGetReal(observer *C.cse_observer) float64 {
	realOutVar := C.uint(0)
	realOutVal := C.double(-1.0)
	C.cse_observer_slave_get_real(observer, 0, &realOutVar, 1, &realOutVal)
	return float64(realOutVal)
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
		observerGetRealSamples(observer, 10, &status.TrendSignals[0])
		setSimulationStatusValue(&status.Module, observer)
		time.Sleep(500 * time.Millisecond)
	}
}
func setSimulationStatusValue(module *Module, observer *C.cse_observer) {
	if len(module.Signals) > 0 {
		module.Signals[0].Value = observerGetReal(observer)
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
				//status.TrendSignals = append(status.TrendSignals, TrendSignal{cmd[1], cmd[2], nil, nil})
			case "untrend":
				status.TrendSignals = []TrendSignal{}
			case "module":
				status.Module = Module{
					Name: cmd[1],
					Signals: []Signal{
						{
							Name:  "Clock",
							Value: -1,
						},
					},
				}
			default:
				fmt.Println("Empty command, mildt sagt not good: ", cmd)
			}
		}
	}
}
