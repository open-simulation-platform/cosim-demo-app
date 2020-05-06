package server

import (
	"cse-server-go/cse"
	"cse-server-go/structs"
	"encoding/json"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func Server(command chan []string, state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *cse.Simulation) {
	router := mux.NewRouter()
	box := packr.NewBox("../resources/public")

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cse.GenerateJsonResponse(simulationStatus, sim, structs.CommandFeedback{}, structs.ShortLivedData{}))
	})

	router.HandleFunc("/modules", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sim.MetaData)
	})

	router.HandleFunc("/plot-config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}).Methods("OPTIONS")

	router.HandleFunc("/plot-config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		configDir := simulationStatus.ConfigDir
		trends := simulationStatus.Trends
		plots := []structs.Plot{}
		for i := 0; i < len(trends); i++ {
			trendValues := trends[i].TrendSignals
 			variables := []structs.PlotVariable{}
			for j := 0; j < len(trendValues); j++ {
				plotVariable := structs.PlotVariable{trendValues[j].Module, trendValues[j].Signal}
				variables = append(variables, plotVariable)
			}
			plot := structs.Plot{trends[i].Label, trends[i].PlotType, variables}
			plots = append(plots, plot)
		}
		plotConfig := structs.PlotConfig{plots}
		plotConfigJson, _ := json.Marshal(plotConfig)
		err := ioutil.WriteFile(configDir + "/" + "PlotConfig.json", plotConfigJson, 0644)
		if err != nil {
			log.Println("Could not write PlotConfig to file, data: ", plotConfig, ", error was:", err)
		}
		json.NewEncoder(w).Encode("Wrote plot configuration to " + configDir + "/" + "PlotConfig.json")
	}).Methods("POST")

	router.HandleFunc("/value/{module}/{cardinality}/{signal}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		module := vars["module"]
		cardinality := vars["cardinality"]
		signal := vars["signal"]
		json.NewEncoder(w).Encode(cse.GetSignalValue(module, cardinality, signal))
	})

	router.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		commandRequest := []string{}
		json.Unmarshal(body, &commandRequest)
		command <- commandRequest
	}).Methods("PUT")

	router.HandleFunc("/ws", WebsocketHandler(command, state))

	//Default handler
	router.PathPrefix("/").Handler(http.FileServer(box))

	log.Fatal(http.ListenAndServe(":8000", router))
}
