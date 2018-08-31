package main

/*
	#include <cse.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
	"time"
)

// UGLY GLOBAL VARIABLE
var lastOutValue = 0.0

func printLastError() () {
	fmt.Printf("Error code %d: %s\n", int(C.cse_last_error_code()), C.GoString(C.cse_last_error_message()))
}

func createExecution() (execution *C.struct_cse_execution_s) {
	execution = C.cse_execution_create(0.0)
	return execution
}

func executionAddfmu(execution *C.struct_cse_execution_s, fmuPath string) (slaveIndex C.int) {
	slaveIndex = C.cse_execution_add_slave_from_fmu((*C.struct_cse_execution_s)(execution), C.CString(fmuPath))
	if slaveIndex < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	return slaveIndex
}

func step(execution *C.struct_cse_execution_s, slaveIndex C.int) (realOutVal C.double) {
	dt := C.double(0.01)
	stepResult := C.cse_execution_step(execution, dt)
	if stepResult < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	realOutVar := C.int(0)
	realOutVal = C.double(-1.0)
	getRealResult := C.cse_execution_slave_get_real(
		execution,
		slaveIndex,
		(*C.uint)(unsafe.Pointer(&realOutVar)),
		1,
		(*C.double)(unsafe.Pointer(&realOutVal)))
	if getRealResult < 0 {
		printLastError()
		C.cse_execution_destroy(execution)
	}
	return realOutVal
}

func simulate(execution *C.struct_cse_execution_s, slaveIndex C.int, command chan string) {
	var status = "pause"
	for {
		select {
		case cmd := <-command:
			fmt.Println(cmd)
			switch cmd {
			case "stop":
				return
			case "pause":
				status = "pause"
			default:
				status = "play"
			}
		default:
			if status == "play" {
				out := step(execution, slaveIndex)
				lastOutValue = float64(out)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
