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

var chat = &Chat{rooms: make(map[string]*Room)}

func main() {
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleClient(chat, conn)
}

func handleClient(chat *Chat, conn *websocket.Conn) {
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			conn.WriteJSON(err)
			return
		}

		resMessage, err := chat.processMessage(msg, conn)
		if err != nil {

		}
		if err := conn.WriteJSON(resMessage); err != nil {
			log.Println(err)
			conn.Close()
			return
		}
	}
}
