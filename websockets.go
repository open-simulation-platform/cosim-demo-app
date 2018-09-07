package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"fmt"
)

var upgrader = websocket.Upgrader{}

type JsonRequest struct {
	Command     string `json:"command,omitempty"`
	Module      string `json:"module,omitempty"`
	Modules     bool   `json:"modules,omitempty"`
	Connections bool   `json:"connections,omitempty"`
}

func commandLoop(command chan string, conn *websocket.Conn) {
	for {
		fmt.Println("Waiting for a message")
		json := JsonRequest{}
		err := conn.ReadJSON(&json)
		if err != nil {
			log.Println("read:", err)
			break
		}
		command <- json.Command
	}
}

func stateLoop(state chan string, conn *websocket.Conn) {
	for {
		fmt.Println("waiting for state")
		latestState := <-state
		fmt.Printf("Received state: %s", state)
		err := conn.WriteMessage(websocket.TextMessage, []byte(latestState))
		fmt.Println("Wrote state to socket")
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func WebsocketHandler(command chan string, state chan string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		//defer conn.Close()
		go commandLoop(command, conn)
		go stateLoop(state, conn)
	}
}