package main

import (
	"encoding/json"
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

func (rm *RoomManager) AddConnectionToRoom(room string, conn *websocket.Conn) {
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

func (room_manager *RoomManager) SendMessageToAll(room_name string, data map[string]interface{}) {
	if _, exists := room_manager.Rooms[room_name]; !exists {
		fmt.Printf("Room '%s' does not exist\n", room_name)
		return
	}

	json_data, json_err := json.Marshal(data)

	if json_err != nil {
		fmt.Println("Error Marshalling data:", json_err)
		return
	}
	for conn := range room_manager.Rooms[room_name] {
		conn.WriteMessage(websocket.TextMessage, json_data)
	}
}

func (room_manager *RoomManager) SendMessageToAllExceptSelf(self *websocket.Conn, room_name string, data map[string]interface{}) {
	if _, exists := room_manager.Rooms[room_name]; !exists {
		fmt.Printf("Room '%s' does not exist\n", room_name)
		return
	}

	json_data, json_err := json.Marshal(data)

	if json_err != nil {
		fmt.Println("Error Marshalling data:", json_err)
		return
	}
	for conn := range room_manager.Rooms[room_name] {
		if conn == self {
			continue
		}
		conn.WriteMessage(websocket.TextMessage, json_data)
	}
}
