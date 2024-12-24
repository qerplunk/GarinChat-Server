package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"qerplunk/garin-chat/middleware"
	"time"

	"github.com/gorilla/websocket"
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

/*
Type of WebSocket server messages.
WsJoin: new user joins a room
WsUserLeave: user leaves a room
WsMessage: user sends a message to a room
*/
const (
	WsJoin      = "join"
	WsUserLeave = "userleave"
	WsMessage   = "message"
)

// Room manager used to store and map room names to WebSocket connections
var roomManager *RoomManager = NewRoomManager()

// The frontend will hard limit users sending messages to once every 2.5 seconds
// But due to network timings the server limits to 1 message/second to avoid kicking good users out
const (
	maxMessagesPerSecond = 1
	slidingWindowSeconds = 1
)

type rateLimiter struct {
	// Track the number of messages received in the current time window
	messageCount int
	// Store the timestamp of the last time the counter 'messageCount' was reset
	// Defines the start of the current time window
	lastReset time.Time
}

// Returns if the user is under the rate limit and allowed to send messages to the server
func (rl *rateLimiter) allowMessage() bool {
	now := time.Now()

	// Reset rate limiter every slidingWindowSeconds
	if now.Sub(rl.lastReset) > slidingWindowSeconds*time.Second {
		rl.messageCount = 0
		rl.lastReset = now
	}

	if rl.messageCount >= maxMessagesPerSecond {
		return false
	}

	// User did not go over rate limit
	rl.messageCount += 1
	return true
}

// Handles a WebSocket connection instance
func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	// The currentName and currentRoom variables are set once a user joins a room
	var currentName string
	var currentRoom string

	rateLimiter := rateLimiter{
		lastReset: time.Now(),
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			// Handle user closing browser and terminating WebSocket connection
			// Handles "userleave" messages
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
					dataToSend := map[string]interface{}{
						"type":       WsUserLeave,
						"user":       currentName,
						"totalUsers": len(roomManager.Rooms[currentRoom]),
					}
					roomManager.SendMessageToAll(currentRoom, dataToSend)
				}
			} else {
				fmt.Println("Error reading message:", err)
			}
			break
		}

		// Update rate limiter here to also limit invalid messages getting received
		if !rateLimiter.allowMessage() {
			fmt.Println("Rate limit exceeded, closing connection")

			if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
				dataToSend := map[string]interface{}{
					"type":       WsUserLeave,
					"user":       currentName,
					"totalUsers": len(roomManager.Rooms[currentRoom]),
				}
				roomManager.SendMessageToAll(currentRoom, dataToSend)
			}

			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("Error unmarshalling message:", err)

			if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
				dataToSend := map[string]interface{}{
					"type":       WsUserLeave,
					"user":       currentName,
					"totalUsers": len(roomManager.Rooms[currentRoom]),
				}
				roomManager.SendMessageToAll(currentRoom, dataToSend)
			}

			break
		}

		// Checks the type of message the user sent to the server
		switch msg.Type {
		case WsJoin:
			currentName = msg.Username
			currentRoom = msg.Room
			fmt.Println("JOIN:", currentName, currentRoom)

			if currentName == "" || currentRoom == "" {
				fmt.Println("Empty values")
				break
			}

			roomManager.AddConnectionToRoom(currentRoom, conn)

			dataToSend := map[string]interface{}{
				"type":       WsJoin,
				"user":       currentName,
				"totalUsers": len(roomManager.Rooms[currentRoom]),
			}
			roomManager.SendMessageToAll(currentRoom, dataToSend)

		case WsUserLeave:
			fmt.Println("USERLEAVE:", currentName)
			break

		case WsMessage:
			fmt.Println("MESSAGE")
			fmt.Printf("\t%s: %s > %s\n", currentRoom, currentName, msg.Message)

			dataToSend := map[string]interface{}{
				"type":    WsMessage,
				"user":    currentName,
				"message": msg.Message,
			}
			roomManager.SendMessageToAllExceptSelf(conn, currentRoom, dataToSend)
		}
	}

	fmt.Println("End of WebSocket session")
}

// The basic HTTP connection, not WebSocket yet
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, connErr := upgrader.Upgrade(w, r, nil)

	if connErr != nil {
		fmt.Println("Error upgrading connection:", connErr)
		return
	}

	go handleConnection(conn)
}

func main() {
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
	if serveErr := http.ListenAndServe(":"+port, nil); serveErr != nil {
		fmt.Println("Error starting server:", serveErr)
	}
}
