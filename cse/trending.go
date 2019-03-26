package cse

import (
	"cse-server-go/structs"
	"log"
	"strconv"
	"time"
)

func generateNextTrendId(status *structs.SimulationStatus) int {
	var maxId = 0
	for _, trend := range status.Trends {
		if trend.Id > maxId {
			maxId = trend.Id
		}
	}

	return maxId + 1
}

func addNewTrend(status *structs.SimulationStatus, plotType string, label string) (bool, string) {
	id := generateNextTrendId(status)

	status.Trends = append(status.Trends, structs.Trend{
		Id:           id,
		PlotType:     plotType,
		Label:        label,
		TrendSignals: []structs.TrendSignal{},
		Spec: structs.TrendSpec{
			Auto:  true,
			Range: 10.0}})
	return true, "Added new trend"
}

func addToTrend(sim *Simulation, status *structs.SimulationStatus, module string, signal string, causality string, valueType string, valueReference string, plotIndex string) (bool, string) {

	idx, err := strconv.Atoi(plotIndex)
	fmu := findFmu(sim.MetaData, module)

	if err != nil {
		message := strCat("Cannot parse plotIndex as integer", plotIndex, ", ", err.Error())
		log.Println(message)
		return false, message
	}

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

	status.Trends[idx].TrendSignals = append(status.Trends[idx].TrendSignals, structs.TrendSignal{
		Module:         module,
		SlaveIndex:     fmu.ExecutionIndex,
		Signal:         signal,
		Causality:      causality,
		Type:           valueType,
		ValueReference: varIndex})

	return true, "Added variable to trend"
}

func setTrendLabel(status *structs.SimulationStatus, trendIndex string, trendLabel string) (bool, string) {
	idx, _ := strconv.Atoi(trendIndex)
	status.Trends[idx].Label = trendLabel
	return true, "Modified trend label"
}

func removeAllFromTrend(sim *Simulation, status *structs.SimulationStatus, trendIndex string) (bool, string) {
	var success = true
	var message = "Removed all variables from trend"

	idx, _ := strconv.Atoi(trendIndex)
	for _, trendSignal := range status.Trends[idx].TrendSignals {
		err := observerStopObserving(sim.TrendObserver, trendSignal.SlaveIndex, trendSignal.Type, trendSignal.ValueReference)
		if err != nil {
			message = strCat("Cannot stop observing variable: ", err.Error())
			success = false
			log.Println("Cannot stop observing", err)
		}
	}
	status.Trends[idx].TrendSignals = []structs.TrendSignal{}
	return success, message
}

func removeTrend(status *structs.SimulationStatus, trendIndex string) (bool, string) {
	idx, _ := strconv.Atoi(trendIndex)

	if len(status.Trends) > 1 {
		status.Trends = append(status.Trends[:idx], status.Trends[idx+1:]...)
	} else {
		status.Trends = []structs.Trend{}
	}

	return true, "Removed trend"
}

func TrendLoop(sim *Simulation, status *structs.SimulationStatus) {
	for {
		for _, trend := range status.Trends {
			switch trend.PlotType {
			case "trend":
				if len(trend.TrendSignals) > 0 {
					for i, _ := range trend.TrendSignals {
						var signal = &trend.TrendSignals[i]
						switch signal.Type {
						case "Real":
							observerGetRealSamples(sim.TrendObserver, signal, trend.Spec)
						}
					}
				}
				break
			case "scatter":
				signalCount := len(trend.TrendSignals)
				if signalCount > 0 {
					for j := 0; (j + 1) < signalCount; j += 2 {
						var signal1 = &trend.TrendSignals[j]
						var signal2 = &trend.TrendSignals[j+1]
						if (signal1.Type == "Real" && signal2.Type == "Real") {
							observerGetRealSynchronizedSamples(sim.TrendObserver, signal1, signal2, trend.Spec)
						}
					}
				}
				break
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
