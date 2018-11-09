package server

import (
	"cse-server-go/cse"
	"cse-server-go/structs"
	"encoding/json"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Server(command chan []string, state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *cse.Simulation) {
	router := mux.NewRouter()
	box := packr.NewBox("../resources/public")

	router.Handle("/", http.FileServer(box))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cse.SimulationStatus(simulationStatus, sim))
	})

	router.HandleFunc("/ws", WebsocketHandler(command, state))

	log.Fatal(http.ListenAndServe(":8000", router))
}
