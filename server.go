package main

import (
	"html/template"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"log"
	"encoding/json"
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

type Simulator struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

func RestTest(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Simulator{ID: "id-1", Name: "Coral", Status: "Completely broken"})
}

func Server() {
	router := mux.NewRouter()
	box := packr.NewBox("./resources/public")
	tmpl := template.Must(template.ParseFiles("layout.html"))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})
	router.HandleFunc("/rest-test", RestTest)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))
	router.HandleFunc("/cse", func(w http.ResponseWriter, r *http.Request) {
		data.CseAnswer = "The meaning of life is " + strconv.Itoa(cse_hello())
		tmpl.Execute(w, data)
	})
	router.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		data.CseAnswer = ""
		tmpl.Execute(w, data)
	})
	log.Fatal(http.ListenAndServe(":8000", router))
}
