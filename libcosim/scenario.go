// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package libcosim

/*
	#include <cosim.h>
*/
import "C"
import (
	"cosim-demo-app/structs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func createScenarioManager() (manipulator *C.cosim_manipulator) {
	manipulator = C.cosim_scenario_manager_create()
	return
}

func isScenarioRunning(manipulator *C.cosim_manipulator) bool {
	intVal := C.cosim_scenario_is_running(manipulator)
	return intVal > 0
}

func doesFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true

	} else if os.IsNotExist(err) {
		return false
	} else {
		log.Println("Well, this is awkward.")
		return false
	}
}

func loadScenario(sim *Simulation, status *structs.SimulationStatus, filename string) (bool, string) {
	pathToFile := filepath.Join(status.ConfigDir, "scenarios", filename)
	if !strings.HasSuffix(pathToFile, "json") {
		return false, "Scenario file must be of type *.json"
	}
	if !doesFileExist(pathToFile) {
		return false, strCat("Can't find file ", pathToFile)
	}
	success := C.cosim_execution_load_scenario(sim.Execution, sim.ScenarioManager, C.CString(pathToFile))
	if success < 0 {
		return false, strCat("Problem loading scenario file: ", lastErrorMessage())
	}
	status.CurrentScenario = filename
	return true, strCat("Successfully loaded scenario ", pathToFile)
}

func abortScenario(manipulator *C.cosim_manipulator) (bool, string) {
	intVal := C.cosim_scenario_abort(manipulator)
	if int(intVal) < 0 {
		return false, strCat("Failed to abort scenario: ", lastErrorMessage())
	} else {
		return true, "Scenario aborted"
	}
}

func parseScenario(status *structs.SimulationStatus, filename string) (interface{}, error) {
	pathToFile := filepath.Join(status.ConfigDir, "scenarios", filename)
	jsonFile, err := os.Open(pathToFile)

	if err != nil {
		log.Println("Can't open file:", pathToFile)
		return "", err
	}

	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}
	var data interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "", err
	}
	return data, nil
}

func findScenarios(status *structs.SimulationStatus) (scenarios []string) {
	folder := filepath.Join(status.ConfigDir, "scenarios")
	info, e := os.Stat(folder)
	if os.IsNotExist(e) {
		fmt.Println("Scenario folder does not exist: ", folder)
		return
	} else if !info.IsDir() {
		fmt.Println("Scenario folder is not a directory: ", folder)
		return
	} else {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".json") {
				scenarios = append(scenarios, f.Name())
			}
		}
	}
	return
}
