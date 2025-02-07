package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type commandFunc func(Message, *Client)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var handlers = map[string]commandFunc{
	"create_room":  chat.createRoom,
	"join_room":    chat.joinRoom,
	"send_message": chat.sendMessage,
	"leave_room":   chat.leaveRoom,
}

var chat = &Chat{rooms: make(map[string]*Room)}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
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
	defer conn.Close()

	client := &Client{conn: conn, mu: sync.Mutex{}, comms: make(chan ResponseMessage)}

	go client.readPump()
	go client.writePump(chat)

	select {}
}
