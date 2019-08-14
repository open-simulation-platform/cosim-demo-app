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
)

func createObserver() (observer *C.cse_observer) {
	observer = C.cse_last_value_observer_create()
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

func createFileObserverFromCfg(logPath string, cfgPath string) (observer *C.cse_observer) {
	observer = C.cse_file_observer_create_from_cfg(C.CString(logPath), C.CString(cfgPath))
	return
}

func executionAddObserver(execution *C.cse_execution, observer *C.cse_observer) {
	C.cse_execution_add_observer(execution, observer)
}

func observerDestroy(observer *C.cse_observer) {
	C.cse_observer_destroy(observer)
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
				Value:     realOutVal[k],
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

func observerGetBooleans(observer *C.cse_observer, variables []structs.Variable, slaveIndex int) (boolSignals []structs.Signal) {
	var boolValueRefs []C.cse_variable_index
	var boolVariables []structs.Variable
	var numBooleans int
	for _, variable := range variables {
		if variable.Type == "Boolean" {
			ref := C.cse_variable_index(variable.ValueReference)
			boolValueRefs = append(boolValueRefs, ref)
			boolVariables = append(boolVariables, variable)
			numBooleans++
		}
	}

	if numBooleans > 0 {
		boolOutVal := make([]C.bool, numBooleans)
		C.cse_observer_slave_get_boolean(observer, C.cse_slave_index(slaveIndex), &boolValueRefs[0], C.size_t(numBooleans), &boolOutVal[0])

		boolSignals = make([]structs.Signal, numBooleans)
		for k := range boolVariables {
			boolSignals[k] = structs.Signal{
				Name:      boolVariables[k].Name,
				Causality: boolVariables[k].Causality,
				Type:      boolVariables[k].Type,
				Value:     bool(boolOutVal[k]),
			}
		}
	}
	return boolSignals
}

func observerGetStrings(observer *C.cse_observer, variables []structs.Variable, slaveIndex int) (stringSignals []structs.Signal) {
	var stringValueRefs []C.cse_variable_index
	var stringVariables []structs.Variable
	var numStrings int
	for _, variable := range variables {
		if variable.Type == "String" {
			ref := C.cse_variable_index(variable.ValueReference)
			stringValueRefs = append(stringValueRefs, ref)
			stringVariables = append(stringVariables, variable)
			numStrings++
		}
	}

	if numStrings > 0 {
		stringOutVal := make([]*C.char, numStrings)
		C.cse_observer_slave_get_string(observer, C.cse_slave_index(slaveIndex), &stringValueRefs[0], C.size_t(numStrings), &stringOutVal[0])

		stringSignals = make([]structs.Signal, numStrings)
		for k := range stringVariables {
			stringSignals[k] = structs.Signal{
				Name:      stringVariables[k].Name,
				Causality: stringVariables[k].Causality,
				Type:      stringVariables[k].Type,
				Value:     C.GoString(stringOutVal[k]),
			}
		}
	}
	return stringSignals
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
	signal.TrendXValues = times
	signal.TrendYValues = trendVals
}

func observerGetRealSynchronizedSamples(observer *C.cse_observer, signal1 *structs.TrendSignal, signal2 *structs.TrendSignal, spec structs.TrendSpec) {
	slaveIndex1 := C.cse_slave_index(signal1.SlaveIndex)
	variableIndex1 := C.cse_variable_index(signal1.ValueReference)

	slaveIndex2 := C.cse_slave_index(signal2.SlaveIndex)
	variableIndex2 := C.cse_variable_index(signal2.ValueReference)

	stepNumbers := make([]C.cse_step_number, 2)
	var success C.int
	if spec.Auto {
		duration := C.cse_duration(spec.Range * 1e9)
		success = C.cse_observer_get_step_numbers_for_duration(observer, slaveIndex1, duration, &stepNumbers[0])
	} else {
		tBegin := C.cse_time_point(spec.Begin * 1e9)
		tEnd := C.cse_time_point(spec.End * 1e9)
		success = C.cse_observer_get_step_numbers(observer, slaveIndex1, tBegin, tEnd, &stepNumbers[0])
	}
	if int(success) < 0 {
		return
	}
	first := stepNumbers[0]
	last := stepNumbers[1]

	numSamples := int(last) - int(first) + 1
	cnSamples := C.size_t(numSamples)
	realOutVal1 := make([]C.double, numSamples)
	realOutVal2 := make([]C.double, numSamples)
	actualNumSamples := C.cse_observer_slave_get_real_synchronized_series(observer, slaveIndex1, variableIndex1, slaveIndex2, variableIndex2, first, cnSamples, &realOutVal1[0], &realOutVal2[0])
	ns := int(actualNumSamples)
	if ns <= 0 {
		return
	}
	trendVals1 := make([]float64, ns)
	trendVals2 := make([]float64, ns)
	for i := 0; i < ns; i++ {
		trendVals1[i] = float64(realOutVal1[i])
		trendVals2[i] = float64(realOutVal2[i])
	}
	signal1.TrendXValues = trendVals1
	signal2.TrendYValues = trendVals2
}

func toVariableType(valueType string) (C.cse_variable_type, error) {
	switch valueType {
	case "Real":
		return C.CSE_VARIABLE_TYPE_REAL, nil
	case "Integer":
		return C.CSE_VARIABLE_TYPE_INTEGER, nil
	case "Boolean":
		return C.CSE_VARIABLE_TYPE_BOOLEAN, nil
	case "String":
		return C.CSE_VARIABLE_TYPE_STRING, nil
	}
	return C.CSE_VARIABLE_TYPE_REAL, errors.New(strCat("Unknown variable type:", valueType))
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
