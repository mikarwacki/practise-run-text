package main

import (
	"log"
)

func (c *Chat) createRoom(msg Message, client *Client, ch chan ResponseMessage) {
	c.mu.Lock()
	if _, ok := c.rooms[msg.Room]; ok {
		c.mu.Unlock()
		log.Println("room already exists")
		ch <- ResponseMessage{Message: "room already exists", Room: msg.Room, Error: true}
		return
	}

	room := NewRoom(msg.Room)

	c.rooms[room.name] = room
	c.mu.Unlock()

	room.mu.Lock()
	room.users[client] = true
	room.mu.Unlock()

	go room.broadcast()

	ch <- ResponseMessage{Message: "room successfully created and joined", Room: room.name}
}

func (c *Chat) joinRoom(msg Message, client *Client, ch chan ResponseMessage) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()

	if !ok {
		ch <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.Lock()
	if _, ok := room.users[client]; ok {
		room.mu.Unlock()
		ch <- ResponseMessage{Message: "you are already in the room", Room: msg.Room, Error: true}
		return
	}
	room.users[client] = true
	room.mu.Unlock()

	ch <- ResponseMessage{Message: "room joined", Room: msg.Room}
}

func (c *Chat) leaveRoom(msg Message, client *Client, ch chan ResponseMessage) {
	c.mu.Lock()
	room, ok := c.rooms[msg.Room]
	c.mu.Unlock()
	if !ok {
		ch <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.Lock()
	if _, ok := room.users[client]; !ok {
		room.mu.Unlock()
		ch <- ResponseMessage{Message: "you are not in the room", Room: msg.Room, Error: true}
		return
	}
	delete(room.users, client)
	room.mu.Unlock()
	ch <- ResponseMessage{Message: "room left", Room: msg.Room}
}

func (c *Chat) sendMessage(msg Message, client *Client, ch chan ResponseMessage) {
	c.mu.RLock()
	room, ok := c.rooms[msg.Room]
	c.mu.RUnlock()
	if !ok {
		ch <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	room.mu.RLock()
	if _, ok := room.users[client]; !ok {
		room.mu.RUnlock()
		ch <- ResponseMessage{Message: "you are not in the room", Room: msg.Room, Error: true}
		return
	}
	room.mu.RUnlock()

	room.messages <- ResponseMessage{Message: msg.Message, Room: msg.Room}
	ch <- ResponseMessage{Message: "message sent", Room: msg.Room}
}

func (c *Chat) deleteRoom(msg Message, client *Client, ch chan ResponseMessage) {
	c.mu.Lock()
	room, ok := c.rooms[msg.Room]
	defer c.mu.Unlock()

	if !ok {
		ch <- ResponseMessage{Message: "room does not exist", Room: msg.Room, Error: true}
		return
	}

	close(room.messages)
	delete(c.rooms, msg.Room)
	ch <- ResponseMessage{Message: "room deleted", Room: msg.Room}
}
