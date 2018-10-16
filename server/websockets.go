package server

import (
	"cse-server-go/structs"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type JsonRequest struct {
	Command     []string `json:"command,omitempty"`
	Module      string   `json:"module,omitempty"`
	Modules     bool     `json:"modules,omitempty"`
	Connections bool     `json:"connections,omitempty"`
}


var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true },}

func commandLoop(command chan []string, conn *websocket.Conn) {
	for {
		json := JsonRequest{}
		err := conn.ReadJSON(&json)
		if err != nil {
			log.Println("read:", err)
			break
		}
		if json.Command != nil {
			command <- json.Command
		}
	}
}

func stateLoop(state chan structs.JsonResponse, conn *websocket.Conn) {
	for {
		latestState := <-state
		err := conn.WriteJSON(latestState)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func WebsocketHandler(command chan []string, state chan structs.JsonResponse) func(w http.ResponseWriter, r *http.Request) {
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
