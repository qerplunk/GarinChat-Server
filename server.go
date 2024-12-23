package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"qerplunk/garin-chat/middleware"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Upgrades HTTP connection to WebSocket connection
// Returns true as a middleware is already used to check for the origin
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message type used for unmarshalling WebSocket message into a variable
type Message struct {
	Type     string `json:"type"`
	Message  string `json:"message,omitempty"`
	Username string `json:"username,omitempty"`
	Room     string `json:"room,omitempty"`
}

// Type of WebSocket server messages
// WS_JOIN: new user joins a room
// WS_USERLEAVE: user leaves a room
// WS_MESSAGE: user sends a message to a room
const (
	WS_JOIN      = "join"
	WS_USERLEAVE = "userleave"
	WS_MESSAGE   = "message"
)

// Room manager used to store and map room names to WebSocket connections
var room_manager *RoomManager = NewRoomManager()

// Handles a WebSocket connection instance
func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	// The currentName and currentRoom variables are set once a user joins a room
	var currentName string
	var currentRoom string

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			// Handle user closing browser and terminating WebSocket connection
			// Handles "userleave" messages
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				if users_left := room_manager.RemoveConnection(currentRoom, conn); users_left {
					data_to_send := map[string]interface{}{
						"type":       WS_USERLEAVE,
						"user":       currentName,
						"totalUsers": len(room_manager.Rooms[currentRoom]),
					}
					room_manager.SendMessageToAll(currentRoom, data_to_send)
				}
			} else {
				fmt.Println("Error reading message:", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("Error unmarshalling message:", err)
			continue
		}

		// Checks the type of message the user sent to the server
		switch msg.Type {
		case WS_JOIN:
			currentName = msg.Username
			currentRoom = msg.Room
			fmt.Println("JOIN:", currentName, currentRoom)

			if currentName == "" || currentRoom == "" {
				fmt.Println("Empty values")
				continue
			}

			room_manager.AddConnectionToRoom(currentRoom, conn)

			data_to_send := map[string]interface{}{
				"type":       WS_JOIN,
				"user":       currentName,
				"totalUsers": len(room_manager.Rooms[currentRoom]),
			}
			room_manager.SendMessageToAll(currentRoom, data_to_send)

		case WS_USERLEAVE:
			fmt.Println("USERLEAVE:", currentName)
			break

		case WS_MESSAGE:
			fmt.Println("MESSAGE")
			fmt.Printf("\t%s: %s > %s\n", currentRoom, currentName, msg.Message)

			data_to_send := map[string]interface{}{
				"type":    WS_MESSAGE,
				"user":    currentName,
				"message": msg.Message,
			}
			room_manager.SendMessageToAllExceptSelf(conn, currentRoom, data_to_send)
		}
	}

	fmt.Println("End of WebSocket session")
}

// The basic HTTP connection, not WebSocket yet
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, conn_err := upgrader.Upgrade(w, r, nil)

	if conn_err != nil {
		fmt.Println("Error upgrading connection:", conn_err)
		return
	}

	go handleConnection(conn)
}

func main() {
	goDotEnv_err := godotenv.Load()
	if goDotEnv_err != nil {
		fmt.Println("Error loading .env file")
	}

	middlewareStack := middleware.CreateStack(
		middleware.JWTCheck(),
		middleware.OriginCheck(),
	)

	http.HandleFunc("/", middlewareStack(handleWebSocket))

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("No port selected")
		return
	}

	fmt.Printf("WebSocket server running on ws://localhost:%s/\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if goDotEnv_err != nil {
		fmt.Println("Error starting server:", err)
	}
}
