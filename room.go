package main

import (
	"log"
	"sync"
)

const broadcastBufferSize = 256

type Room struct {
	users    map[*Client]bool
	name     string
	messages chan ResponseMessage
	mu       sync.RWMutex
}

func (r *Room) broadcast() {
	for msg := range r.messages {
		log.Printf("broadcast started, new message: %s", msg.Message)
		r.mu.RLock()
		for client := range r.users {
			go func(client *Client, msg ResponseMessage) {
				client.mu.Lock()
				if err := client.conn.WriteJSON(msg); err != nil {
					log.Println("Error sending message removing client")
					r.mu.Lock()
					delete(r.users, client)
					r.mu.Unlock()
				}
				client.mu.Unlock()
			}(client, msg)
		}
		r.mu.RUnlock()
	}
	log.Println("broadcast ended")
}

func NewRoom(name string) *Room {
	return &Room{
		name:     name,
		users:    make(map[*Client]bool),
		messages: make(chan ResponseMessage, broadcastBufferSize),
	}
}
