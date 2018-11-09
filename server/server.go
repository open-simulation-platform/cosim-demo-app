package server

import (
	"cse-server-go/cse"
	"cse-server-go/structs"
	"encoding/json"
	"html/template"
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"github.com/gobuffalo/packr"
)

type PageData struct {
	PageTitle string
	CseAnswer string
}

var data = PageData{
	PageTitle: "CSE Hello World",
	CseAnswer: "",
}

func Server(command chan []string, state chan structs.JsonResponse, simulationStatus *structs.SimulationStatus, sim *cse.Simulation) {
	router := mux.NewRouter()
	box := packr.NewBox("../resources/public")
	tmpl := template.Must(template.ParseFiles("../layout.html"))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cse.SimulationStatus(simulationStatus, sim))
	})

	router.HandleFunc("/ws", WebsocketHandler(command, state))

	log.Fatal(http.ListenAndServe(":8000", router))
}
