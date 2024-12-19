package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type RoomManager struct {
	Rooms map[string]map[*websocket.Conn]bool
	mu    sync.Mutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

func (rm *RoomManager) AddConnection(room string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.Rooms[room]; !exists {
		rm.Rooms[room] = make(map[*websocket.Conn]bool)
	}

	rm.Rooms[room][conn] = true
}

func (rm *RoomManager) RemoveConnection(room string, conn *websocket.Conn) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	users_left := true

	if conns, exists := rm.Rooms[room]; exists {
		delete(conns, conn)

		if len(conns) == 0 {
			delete(rm.Rooms, room)
			fmt.Printf("Room '%s' is empty, closing...\n", room)
			users_left = false
		} 
	}
	return users_left
}
