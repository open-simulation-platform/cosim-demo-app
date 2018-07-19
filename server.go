package main

import (
	"html/template"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"log"
)

type PageData struct {
	PageTitle string
	CseAnswer string
}

var data = PageData{
	PageTitle: "CSE Hello World",
	CseAnswer: "",
}

func Server() {
	router := mux.NewRouter()
	tmpl := template.Must(template.ParseFiles("layout.html"))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./resources/public"))))
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
