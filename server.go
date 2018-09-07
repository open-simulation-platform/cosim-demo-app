package main

import (
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

func Server(command chan string, state chan string) {
	router := mux.NewRouter()
	box := packr.NewBox("./resources/public")
	tmpl := template.Must(template.ParseFiles("layout.html"))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	router.HandleFunc("/ws", WebsocketHandler(command, state))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	log.Fatal(http.ListenAndServe(":8000", router))
}
