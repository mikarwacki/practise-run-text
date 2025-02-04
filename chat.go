package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

const broadcastBufferSize = 256

type Chat struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

type Message struct {
	Message string `json:"message"`
	Room    string `json:"room"`
	Command string `json:"command"`
}

type ResponseMessage struct {
	Room    string `json:"room,omitempty"`
	Message string `json:"message"`
}

type Room struct {
	users    map[*websocket.Conn]bool
	name     string
	messages chan ResponseMessage
	mu       sync.RWMutex
}

func (r *Room) broadcast() {
	for msg := range r.messages {
		r.mu.RLock()
		for conn := range r.users {
			go func(conn *websocket.Conn, msg ResponseMessage) {
				if err := conn.WriteJSON(msg); err != nil {
					r.mu.Lock()
					delete(r.users, conn)
					r.mu.Unlock()
				}
			}(conn, msg)
		}
		r.mu.RUnlock()
	}
}

func NewRoom(name string) *Room {
	return &Room{
		name:     name,
		users:    make(map[*websocket.Conn]bool),
		messages: make(chan ResponseMessage, broadcastBufferSize),
	}
}

func (c *Chat) processMessage(msg Message, conn *websocket.Conn) (ResponseMessage, error) {
	switch msg.Command {
	case "create_room":
		return c.createRoom(msg, conn)
	case "join_room":
		return c.joinRoom(msg, conn)
	case "leave_room":
		return c.leaveRoom(msg, conn)
	case "send_message":
		return c.sendMessage(msg, conn)
	case "delete_room":
		return c.deleteRoom(msg)
	default:
		return ResponseMessage{}, fmt.Errorf("Invalid command")
	}
}
