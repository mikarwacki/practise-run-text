package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func (c *Chat) createRoom(msg Message, conn *websocket.Conn) (ResponseMessage, error) {
	c.mu.Lock()
	if _, ok := c.rooms[msg.Room]; ok {
		c.mu.Unlock()
		log.Println("room already exists")
		return ResponseMessage{}, fmt.Errorf("Room %s already exists", msg.Room)
	}

	room := NewRoom(msg.Room)

	c.rooms[room.name] = room
	c.mu.Unlock()

	room.mu.Lock()
	room.users[conn] = true
	room.mu.Unlock()

	go room.broadcast()

	return ResponseMessage{Message: "Room successfully created and joined", Room: room.name}, nil
}

func (c *Chat) joinRoom(msg Message, conn *websocket.Conn) (ResponseMessage, error) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()

	if !ok {
		return ResponseMessage{}, fmt.Errorf("Room %s does not exist", msg.Room)
	}

	room.mu.Lock()
	room.users[conn] = true
	room.mu.Unlock()

	return ResponseMessage{Message: "Room joined", Room: msg.Room}, nil
}

func (c *Chat) leaveRoom(msg Message, conn *websocket.Conn) (ResponseMessage, error) {
	c.mu.Lock()
	room, ok := c.rooms[msg.Room]
	c.mu.Unlock()
	if !ok {
		return ResponseMessage{}, fmt.Errorf("Room %s does not exist", msg.Room)
	}

	room.mu.Lock()
	delete(room.users, conn)
	room.mu.Unlock()
	return ResponseMessage{Message: "Room left", Room: msg.Room}, nil
}

func (c *Chat) sendMessage(msg Message, conn *websocket.Conn) (ResponseMessage, error) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()
	if !ok {
		return ResponseMessage{}, fmt.Errorf("Room %s does not exist", msg.Room)
	}
	room.mu.RLock()

	if _, ok := room.users[conn]; !ok {
		return ResponseMessage{}, fmt.Errorf("You are not in the room")
	}

	room.messages <- ResponseMessage{Message: msg.Message, Room: msg.Room}
	return ResponseMessage{Message: "Message sent", Room: msg.Room}, nil
}

func (c *Chat) deleteRoom(msg Message) (ResponseMessage, error) {
	c.mu.Lock()
	room, ok := c.rooms[msg.Room]

	if !ok {
		c.mu.Unlock()
		return ResponseMessage{}, fmt.Errorf("Room %s does not exist", msg.Room)
	}

	close(room.messages)
	delete(c.rooms, msg.Room)
	return ResponseMessage{Message: "Room deleted", Room: msg.Room}, nil
}
