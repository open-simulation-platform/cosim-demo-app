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

type JsonResponse struct {
	Modules     []string `json:"modules,omitempty"`
	Status      string   `json:"status,omitempty"`
	SignalValue float64  `json:"signalValue,omitempty"`
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
		if json.Command != "" {
			command <- json.Command
		}
		conn.WriteJSON(
			JsonResponse{
				Modules:     []string{"Clock"},
				SignalValue: lastOutValue,
				Status:      "running",
			})
	}
}

func stateLoop(state chan string, conn *websocket.Conn) {
	for {
		latestState := <-state
		err := conn.WriteMessage(websocket.TextMessage, []byte(latestState))
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