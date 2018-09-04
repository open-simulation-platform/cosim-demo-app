package main

import (
	"html/template"
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"encoding/json"
	"github.com/gobuffalo/packr"
	"fmt"
)

type PageData struct {
	PageTitle string
	CseAnswer string
}

var data = PageData{
	PageTitle: "CSE Hello World",
	CseAnswer: "",
}

type Simulator struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status,omitempty"`
	SignalValue string `json:"signalValue,omitempty"`
}

func Server(command chan string, state chan string) {
	router := mux.NewRouter()
	box := packr.NewBox("./resources/public")
	tmpl := template.Must(template.ParseFiles("layout.html"))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})

	router.HandleFunc("/rest-test", func(w http.ResponseWriter, r *http.Request) {
		sigVal := fmt.Sprintf("%.2f", lastOutValue)
		json.NewEncoder(w).Encode(Simulator{ID: "id-1", Name: "Clock", Status: "-", SignalValue: sigVal})
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	router.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		command <- "play"
		json.NewEncoder(w).Encode(Simulator{ID: "id-1", Name: "Clock", Status: "-", SignalValue: "1.0"})
	})

	router.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		command <- "pause"
		json.NewEncoder(w).Encode(Simulator{ID: "id-1", Name: "Clock", Status: "-"})
	})
	router.HandleFunc("/ws", WebsocketHandler(command, state))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	log.Fatal(http.ListenAndServe(":8000", router))
}
