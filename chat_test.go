package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func createTestClient(t *testing.T, server *httptest.Server) (*websocket.Conn, error) {
	wsURl := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURl, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection: %v", err)
	}
	return conn, err
}

func sendMessage(conn *websocket.Conn, command, room, message string) (*ResponseMessage, error) {
	msg := Message{
		Command: command,
		Room:    room,
		Message: message,
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		return nil, err
	}

	var response ResponseMessage
	err = conn.ReadJSON(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func TestCreateRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	}))
	defer func() {
		server.Close()
		chat.rooms = make(map[string]*Room)
	}()

	conn, err := createTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tests := []struct {
		command string
		room    string
		error   bool
	}{
		{
			command: "create_room",
			room:    "general",
			error:   false,
		},
		{
			command: "create_room",
			room:    "general",
			error:   true,
		},
	}

	for _, test := range tests {
		response, err := sendMessage(conn, test.command, test.room, "")
		if err != nil {
			t.Fatalf("could not send message: %v", err)
		}
		if response.Error != test.error {
			t.Fatalf("expected error: %v, got: %v", test.error, response.Error)
		}
	}
}

func TestJoinRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	}))
	defer func() {
		server.Close()
		chat.rooms = make(map[string]*Room)
	}()
	conn, err := createTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tests := []struct {
		command      string
		room         string
		errorMessage string
		error        bool
	}{
		{
			command: "join_room",
			room:    "general",
			error:   true,
		},
		{
			command: "create_room",
			room:    "general",
			error:   false,
		},
		{
			command:      "join_room",
			room:         "general",
			errorMessage: "you are already in the room",
			error:        true,
		},
	}

	for _, test := range tests {
		response, err := sendMessage(conn, test.command, test.room, "")
		if err != nil {
			t.Fatalf("could not send message: %v", err)
		}
		if response.Error != test.error && response.Message != test.errorMessage {
			t.Fatalf("expected error: %v, got: %v", test.error, response.Error)
		}
	}
}

func TestLeaveRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	}))
	defer func() {
		server.Close()
		chat.rooms = make(map[string]*Room)
	}()

	conn, err := createTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tests := []struct {
		command string
		room    string
		error   bool
	}{
		{
			command: "leave_room",
			room:    "general",
			error:   true,
		},
		{
			command: "create_room",
			room:    "general",
			error:   false,
		},
		{
			command: "leave_room",
			room:    "general",
			error:   false,
		},
	}

	for _, test := range tests {
		response, err := sendMessage(conn, test.command, test.room, "")
		if err != nil {
			t.Fatalf("could not send message: %v", err)
		}
		if response.Error != test.error {
			t.Fatalf("expected error: %v, got: %v", test.error, response.Error)
		}
	}
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	}))
	defer func() {
		server.Close()
		chat.rooms = make(map[string]*Room)
	}()

	conn, err := createTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	conn2, err := createTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	tests := []struct {
		command string
		room    string
		message string
		error   bool
	}{
		{
			command: "create_room",
			room:    "general",
			error:   false,
		},
		{
			command: "send_message",
			room:    "general",
			message: "Hello general",
			error:   false,
		},
	}

	for _, test := range tests {
		response, err := sendMessage(conn, test.command, test.room, test.message)
		if err != nil {
			t.Fatalf("could not send message: %v", err)
		}
		if response.Error != test.error {
			t.Fatalf("expected error: %v, got: %v", test.error, response.Error)
		}
		if response.Message == "message sent" {
			var response2 ResponseMessage
			err = conn2.ReadJSON(&response2)
			if err != nil {
				t.Fatal(err)
			}
			if response2.Message != test.message {
				t.Fatalf("expected message: %s, got: %s", test.message, response2.Message)
			}
		}
	}
}
