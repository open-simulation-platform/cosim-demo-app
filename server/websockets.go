package server

import (
	"cse-server-go/structs"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
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
		data := JsonRequest{}
		_, r, err := conn.NextReader()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		err = json.NewDecoder(r).Decode(&data)
		if err == io.EOF {
			// One value is expected in the message.
			log.Println("Message was EOF:", err, data)
			err = io.ErrUnexpectedEOF
		} else if err != nil {
			log.Println("Could not parse message:", data, ", error was:", err)
		} else if data.Command != nil {
			command <- data.Command
		}
	}
}

func stateLoop(state chan structs.JsonResponse, conn *websocket.Conn) {
	for {
		latestState := <-state
		err := conn.WriteJSON(latestState)
		if err != nil {
			log.Println("write:", err)
			log.Println("latestState:", latestState)
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
