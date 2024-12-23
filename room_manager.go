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

func (roomManager *RoomManager) AddConnectionToRoom(room string, conn *websocket.Conn) {
	roomManager.mu.Lock()
	defer roomManager.mu.Unlock()

	if _, exists := roomManager.Rooms[room]; !exists {
		roomManager.Rooms[room] = make(map[*websocket.Conn]bool)
	}

	roomManager.Rooms[room][conn] = true
}

func (roomManager *RoomManager) RemoveConnection(room string, conn *websocket.Conn) bool {
	roomManager.mu.Lock()
	defer roomManager.mu.Unlock()

	usersLeft := true

	if conns, exists := roomManager.Rooms[room]; exists {
		delete(conns, conn)

		if len(conns) == 0 {
			delete(roomManager.Rooms, room)
			fmt.Printf("Room '%s' is empty, closing...\n", room)
			usersLeft = false
		} 
	}
	return usersLeft
}

func (roomManager *RoomManager) SendMessageToAll(roomName string, data map[string]interface{}) {
	if _, exists := roomManager.Rooms[roomName]; !exists {
		fmt.Printf("Room '%s' does not exist\n", roomName)
		return
	}

	jsonData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		fmt.Println("Error Marshalling data:", jsonErr)
		return
	}
	for conn := range roomManager.Rooms[roomName] {
		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}

func (roomManager *RoomManager) SendMessageToAllExceptSelf(self *websocket.Conn, roomName string, data map[string]interface{}) {
	if _, exists := roomManager.Rooms[roomName]; !exists {
		fmt.Printf("Room '%s' does not exist\n", roomName)
		return
	}

	jsonData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		fmt.Println("Error Marshalling data:", jsonErr)
		return
	}
	for conn := range roomManager.Rooms[roomName] {
		if conn == self {
			continue
		}
		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}
