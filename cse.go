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

func simulate(execution *C.cse_execution, observer *C.cse_observer, command chan string) {
	var status = "pause"
	for {
		select {
		case cmd := <-command:
			switch cmd {
			case "stop":
				return
			case "pause":
				executionStop(execution)
				status = "pause"
			default:
				executionStart(execution)
				status = "play"
			}
		default:
			if status == "play" {
				status := executionGetStatus(execution)
				fmt.Println("Current time: ", status.current_time)
				lastOutValue = observerGetReal(observer)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
