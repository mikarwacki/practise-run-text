package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type messageHandler func(Message, *websocket.Conn) (ResponseMessage, error)

var handlers = map[string]messageHandler{
	"create_room":  chat.createRoom,
	"join_room":    chat.joinRoom,
	"send_message": chat.sendMessage,
	"leave_room":   chat.leaveRoom,
}

var chat = &Chat{rooms: make(map[string]*Room)}

func main() {
	http.HandleFunc("/ws", serveWs)
	log.Println("server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	log.Printf("new client connected: %v", r.RemoteAddr)
	if err != nil {
		log.Println(err)
		return
	}

	go handleClient(conn)
}

func handleClient(conn *websocket.Conn) {
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		if handler, ok := handlers[msg.Command]; ok {
			log.Printf("handling msg command: %s", msg.Command)
			resMessage, err := handler(msg, conn)

			if err != nil {
				type ErrorMessage struct {
					Error string `json:"error"`
				}
				if err := conn.WriteJSON(ErrorMessage{Error: err.Error()}); err != nil {
					log.Println(err)
					return
				}
			} else {
				if err := conn.WriteJSON(resMessage); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
