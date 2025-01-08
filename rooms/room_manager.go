package rooms

import (
	"encoding/json"
	"log"
	"qerplunk/garin-chat/types"

	"github.com/gorilla/websocket"
)

// Room service type for NewRoomService.
type RoomService struct {
	*types.RoomManager
}

// Create a new room service and uses the RoomService type.
// Avoids "cannot define new methods on non-local type types.RoomManager" error.
// Does not return "*types.RoomManager" since it cannot be used as a receiver.
// Usage: var roomManager *rooms.RoomService = rooms.NewRoomService().
func NewRoomService() *RoomService {
	return &RoomService{
		RoomManager: types.NewRoomManager(),
	}
}

// Adds a connection to a room.
// Creates a new map if the room does not exist.
func (roomManager RoomService) AddConnectionToRoom(room string, conn *websocket.Conn) {
	roomManager.Mu.Lock()
	defer roomManager.Mu.Unlock()

	if _, exists := roomManager.Rooms[room]; !exists {
		roomManager.Rooms[room] = make(map[*websocket.Conn]bool)
	}

	roomManager.Rooms[room][conn] = true
}

// Removes a connection from a room.
// If there are no more users left in the room, it deletes the map entry and returns false for no more users left.
// Otherwise, it returns true since there are users left in the room.
func (roomManager RoomService) RemoveConnection(room string, conn *websocket.Conn) bool {
	roomManager.Mu.Lock()
	defer roomManager.Mu.Unlock()

	usersLeft := true

	if conns, exists := roomManager.Rooms[room]; exists {
		delete(conns, conn)

		if len(conns) == 0 {
			delete(roomManager.Rooms, room)
			log.Printf("Room '%s' is empty, closing\n", room)
			usersLeft = false
		}
	}
	return usersLeft
}

// Sends a message to all users within the room.
func (roomManager RoomService) SendMessageToAll(roomName string, data types.Message) {
	if _, exists := roomManager.Rooms[roomName]; !exists {
		log.Printf("Room '%s' does not exist\n", roomName)
		return
	}

	jsonData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		log.Println("Error Marshalling data:", jsonErr)
		return
	}
	for conn := range roomManager.Rooms[roomName] {
		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}

// Sends a message to all users (except the sender) within the room.
func (roomManager RoomService) SendMessageToAllExceptSelf(self *websocket.Conn, roomName string, data types.Message) {
	if _, exists := roomManager.Rooms[roomName]; !exists {
		log.Printf("Room '%s' does not exist\n", roomName)
		return
	}

	jsonData, jsonErr := json.Marshal(data)

	if jsonErr != nil {
		log.Println("Error Marshalling data:", jsonErr)
		return
	}
	for conn := range roomManager.Rooms[roomName] {
		if conn == self {
			continue
		}
		conn.WriteMessage(websocket.TextMessage, jsonData)
	}
}
