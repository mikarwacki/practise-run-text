package main

import (
	"log"
)

func (c *Chat) createRoom(msg Message, client *Client) {
	c.mu.Lock()
	if _, ok := c.rooms[msg.Room]; ok {
		c.mu.Unlock()
		log.Println("room already exists")
		client.comms <- ResponseMessage{Message: "room already exists", Room: msg.Room, Error: true}
		return
	}

	room := NewRoom(msg.Room)

	c.rooms[room.name] = room
	c.mu.Unlock()

	room.mu.Lock()
	room.users[client] = true
	room.mu.Unlock()

	client.comms <- ResponseMessage{Message: "room successfully created and joined", Room: room.name}
}

func (c *Chat) joinRoom(msg Message, client *Client) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()

	if !ok {
		client.comms <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.Lock()
	if _, ok := room.users[client]; ok {
		room.mu.Unlock()
		client.comms <- ResponseMessage{Message: "you are already in the room", Room: msg.Room, Error: true}
		return
	}
	room.users[client] = true
	room.mu.Unlock()

	client.comms <- ResponseMessage{Message: "room joined", Room: msg.Room}
}

func (c *Chat) leaveRoom(msg Message, client *Client) {
	c.mu.Lock()
	room, ok := c.rooms[msg.Room]
	c.mu.Unlock()
	if !ok {
		client.comms <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.Lock()
	if _, ok := room.users[client]; !ok {
		room.mu.Unlock()
		client.comms <- ResponseMessage{Message: "you are not in the room", Room: msg.Room, Error: true}
		return
	}
	delete(room.users, client)
	room.mu.Unlock()
	client.comms <- ResponseMessage{Message: "room left", Room: msg.Room}
}

func (c *Chat) sendMessage(msg Message, client *Client) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()
	if !ok {
		client.comms <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.RLock()
	if _, ok := room.users[client]; !ok {
		room.mu.RUnlock()
		client.comms <- ResponseMessage{Message: "you are not in the room", Room: msg.Room, Error: true}
		return
	}
	room.mu.RUnlock()

	for client := range room.users {
		client.comms <- ResponseMessage{Message: msg.Message, Room: msg.Room}
	}
	client.comms <- ResponseMessage{Message: "message sent", Room: msg.Room}
}

