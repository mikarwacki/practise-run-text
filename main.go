package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type messageHandler func(Message, *Client)

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
	defer close(client.comms)

	go client.readPump()
	go client.writePump(chat)

	select {}
}

func (c *Client) readPump() {
	log.Println("read pump started")
	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		if handler, ok := handlers[msg.Command]; ok {
			log.Printf("handling msg command: %s", msg.Command)
			go handler(msg, c)
		} else {
			log.Printf("unknown command: %s", msg.Command)
			c.comms <- ResponseMessage{Message: "unknown command"}
		}
	}
}

func (c *Client) writePump(chat *Chat) {
	log.Println("write pump started")
	defer cleanUp(c, chat)
	for {
		select {
		case msg, ok := <-c.comms:
			if !ok {
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Println("Error sending message ")
				return
			}
		}
	}
}

func cleanUp(client *Client, chat *Chat) {
	log.Println("cleaning up a connection")
	chat.mu.RLock()
	defer chat.mu.RUnlock()
	for _, room := range chat.rooms {
		if _, ok := room.users[client]; ok {
			room.mu.Lock()
			delete(room.users, client)
			room.mu.Unlock()
		}
	}
}
