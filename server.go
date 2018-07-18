package main

import (
	"html/template"
	"net/http"
	"strconv"
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
	tmpl := template.Must(template.ParseFiles("layout.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	})
	fs := http.FileServer(http.Dir("./resources/public"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	http.HandleFunc("/cse", func(w http.ResponseWriter, r *http.Request) {
		data.CseAnswer = "The meaning of life is " + strconv.Itoa(cse_hello())
		tmpl.Execute(w, data)
	})
	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		data.CseAnswer = ""
		tmpl.Execute(w, data)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
