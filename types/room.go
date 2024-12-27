package types

import (
	"sync"

	"github.com/gorilla/websocket"
)

// RoomManager data structure type used for creating new room managers and services.
type RoomManager struct {
	Rooms map[string]map[*websocket.Conn]bool
	Mu    sync.Mutex
}

// Keep this here as NewRoomManager does not have any higher-level logic and is tightly
// coupled with RoomManager struct.
func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]map[*websocket.Conn]bool),
	}
}
