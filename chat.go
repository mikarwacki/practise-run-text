package main

import (
	"sync"
)

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
	Error   bool   `json:"error,omitempty"`
}
