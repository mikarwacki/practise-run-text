package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type messageHandler func(Message, *Client, chan ResponseMessage)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

	client := &Client{conn: conn, mu: sync.Mutex{}}
	for {
		var msg Message
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		if handler, ok := handlers[msg.Command]; ok {
			log.Printf("handling msg command: %s", msg.Command)
			ch := make(chan ResponseMessage)
			go handler(msg, client, ch)

			resMessage := <-ch
			client.mu.Lock()
			if err := conn.WriteJSON(resMessage); err != nil {
				client.mu.Unlock()
				log.Println(err)
				return
			}
			client.mu.Unlock()
		} else {
			log.Printf("unknown command: %s", msg.Command)
			resMessage := ResponseMessage{Message: "unknown command"}
			if err := conn.WriteJSON(resMessage); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
