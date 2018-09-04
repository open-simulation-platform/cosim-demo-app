package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
)

var upgrader = websocket.Upgrader{}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	conn.WriteMessage(websocket.TextMessage, []byte("{\"Hello\": \"Mordi\"}"))
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}