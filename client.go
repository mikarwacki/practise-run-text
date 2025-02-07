package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeLimit = 10 * time.Second
	pingPeriod = 50 * time.Second
	pongWait   = 60 * time.Second
)

type Client struct {
	conn  *websocket.Conn
	comms chan ResponseMessage
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

		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		cleanUp(c, chat)
	}()
	for {
		select {
		case msg, ok := <-c.comms:
			if !ok {
				log.Println("channel comms closed")
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeLimit))
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeLimit))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
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
