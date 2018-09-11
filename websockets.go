package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"fmt"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true },}

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

func stateLoop(state chan JsonResponse, conn *websocket.Conn) {
	for {
		latestState := <-state
		err := conn.WriteJSON(latestState)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func WebsocketHandler(command chan string, state chan JsonResponse) func(w http.ResponseWriter, r *http.Request) {
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
