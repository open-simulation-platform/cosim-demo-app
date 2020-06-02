// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package libcosim

import (
	"cosim-demo-app/structs"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
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

func addToTrend(sim *Simulation, status *structs.SimulationStatus, module string, signal string, plotIndex string) (bool, string) {

	idx, err := strconv.Atoi(plotIndex)
	if err != nil {
		message := strCat("Cannot parse plotIndex as integer", plotIndex, ", ", err.Error())
		log.Println(message)
		return false, message
	}

	fmu, err := findFmu(sim.MetaData, module)
	if err != nil {
		message := err.Error()
		log.Println(message)
		return false, message
	}

	variable, err := findVariable(fmu, signal)
	if err != nil {
		message := err.Error()
		log.Println(message)
		return false, message
	}

	err = observerStartObserving(sim.TrendObserver, fmu.ExecutionIndex, variable.Type, variable.ValueReference)
	if err != nil {
		message := strCat("Cannot start observing variable ", lastErrorMessage())
		log.Println(message)
		return false, message
	}

	status.Trends[idx].TrendSignals = append(status.Trends[idx].TrendSignals, structs.TrendSignal{
		Module:         module,
		SlaveIndex:     fmu.ExecutionIndex,
		Signal:         signal,
		Causality:      variable.Causality,
		Type:           variable.Type,
		ValueReference: variable.ValueReference})

	return true, "Added variable to trend"
}

func setTrendLabel(status *structs.SimulationStatus, trendIndex string, trendLabel string) (bool, string) {
	idx, _ := strconv.Atoi(trendIndex)
	var uuid = rand.Intn(9999)*rand.Intn(9999) + rand.Intn(9999)
	status.Trends[idx].Label = strCat(trendLabel, " #", strconv.Itoa(uuid))
	return true, "Modified trend label"
}

func removeAllFromTrend(sim *Simulation, status *structs.SimulationStatus, trendIndex string) (bool, string) {
	var success = true
	var message = "Removed all variables from trend"

	idx, _ := strconv.Atoi(trendIndex)
	for _, trendSignal := range status.Trends[idx].TrendSignals {
		err := observerStopObserving(sim.TrendObserver, trendSignal.SlaveIndex, trendSignal.Type, trendSignal.ValueReference)
		if err != nil {
			message = strCat("Cannot stop observing variable: ", lastErrorMessage())
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

func activeTrend(status *structs.SimulationStatus, trendIndex string) (bool, string) {
	if len(trendIndex) > 0 {
		idx, err := strconv.Atoi(trendIndex)
		if err != nil {
			return false, strCat("Could not parse trend index: ", trendIndex, " ", err.Error())
		}
		status.ActiveTrend = idx
	} else {
		status.ActiveTrend = -1
	}
	for _, trend := range status.Trends {
		for i, _ := range trend.TrendSignals {
			trend.TrendSignals[i].TrendXValues = nil
			trend.TrendSignals[i].TrendYValues = nil
		}
	}
	return true, "Changed active trend index"
}

func TrendLoop(sim *Simulation, status *structs.SimulationStatus) {
	for {
		for _, trend := range status.Trends {
			if status.ActiveTrend != trend.Id {
				continue
			}
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
						if signal1.Type == "Real" && signal2.Type == "Real" {
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

func parsePlotConfig(pathToFile string) (data structs.PlotConfig, err error) {
	jsonFile, err := os.Open(pathToFile)

	if err != nil {
		log.Println("Can't open file:", err.Error(), pathToFile)
		return data, err
	}

	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println("Can't read file:", err.Error(), pathToFile)
		return data, err
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		log.Println("Can't unmarshal PlotConfig json contents:", err.Error())
		return data, err
	}
	return data, nil
}

func setupPlotsFromConfig(sim *Simulation, status *structs.SimulationStatus, configDir string) {
	if hasFile(configDir, "PlotConfig.json") {
		log.Println("We have a PlotConfig")
		pathToFile := filepath.Join(configDir, "PlotConfig.json")
		plotConfig, err := parsePlotConfig(pathToFile)

		if err != nil {
			log.Println("Can't parse PlotConfig.json:", err.Error())
			return
		}

		for idx, plot := range plotConfig.Plots {
			success, message := addNewTrend(status, plot.PlotType, plot.Label)
			if !success {
				log.Println("Could not add new plot:", message)
			} else {
				plotIdx := strconv.Itoa(idx)
				for _, variable := range plot.PlotVariables {
					success, message := addToTrend(sim, status, variable.Simulator, variable.Variable, plotIdx)
					if !success {
						log.Println("Could not add variable to plot:", message)
					}
				}
			}
		}

	}
}
