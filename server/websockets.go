package server

import (
	"cse-server-go/structs"
	"github.com/gorilla/websocket"
	"github.com/ugorji/go/codec"
	"io"
	"log"
	"net/http"
	"reflect"
)

type JsonRequest struct {
	Command     []string `json:"command,omitempty"`
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true },}

func commandLoop(command chan []string, conn *websocket.Conn) {
	var (
		mh codec.MsgpackHandle
	)
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoder(nil, &mh)

	for {
		data := JsonRequest{}
		_, r, err := conn.NextReader()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		decoder.Reset(r)
		err = decoder.Decode(&data)

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
	var (
		mh codec.MsgpackHandle
	)
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	encoder := codec.NewEncoder(nil, &mh)
	for {
		latestState := <-state

		w, err := conn.NextWriter(2)
		encoder.Reset(w)
		err = encoder.Encode(latestState)
		if err != nil {
			log.Println("write error:", err)
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
