package main

import (
	"sync"
)

type Room struct {
	users map[*Client]bool
	name  string
	mu    sync.RWMutex
}

func NewRoom(name string) *Room {
	return &Room{
		name:  name,
		users: make(map[*Client]bool),
	}
}
