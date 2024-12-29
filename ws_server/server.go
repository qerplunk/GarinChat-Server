package wsserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qerplunk/garin-chat/auth"
	"qerplunk/garin-chat/rooms"
	"qerplunk/garin-chat/types"
	"qerplunk/garin-chat/ws_server/rate_limiter"
	"sync"
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

// Type of WebSocket server messages.
// WsJoin: new user joins a room
// WsUserLeave: user leaves a room
// WsMessage: user sends a message to a room
const (
	WsAuth      = "auth"
	WsJoin      = "join"
	WsUserLeave = "userleave"
	WsMessage   = "message"
)

// Room manager service used to store and map room names to WebSocket connections
var roomManager *rooms.RoomService = rooms.NewRoomService()

// Handles a WebSocket connection instance
func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	// The currentName and currentRoom variables are set once a user joins a room
	var currentName, currentRoom string
	var authenticated = false
	var hasJoined = false

	rateLimiter := ratelimiter.RateLimiter{
		LastReset: time.Now(),
	}

	// Auth and join timeout channels to avoid users from:
	// 1. Connecting without authenticating
	// 2. Authenticating without joining room
	authTimeoutChannel := make(chan struct{})
	joinTimeoutChannel := make(chan struct{})

	// Use a mutex since multiple places can close the connection
	var closeMutex sync.Mutex
	var isConnClosed bool = false

	// Local goroutine to handle channels
	// User has to auth AND join to avoid timeout
	go func() {
		select {
		case <-authTimeoutChannel:
			go func() {
				select {
				case <-joinTimeoutChannel:
					return
				case <-time.After(2 * time.Second):
					fmt.Println("Join timeout")
					closeMutex.Lock()
					defer closeMutex.Unlock()
					isConnClosed = true
					conn.Close()
				}
			}()
		case <-time.After(2 * time.Second):
			fmt.Println("Auth timeout")
			closeMutex.Lock()
			defer closeMutex.Unlock()
			isConnClosed = true
			conn.Close()
		}
	}()

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			// Return since conn.Close() will immediately make websocket.IsCloseError to return an error
			// Since the user has not joined we don't have to handle anything else
			closeMutex.Lock()
			defer closeMutex.Unlock()
			if isConnClosed {
				fmt.Println("Connection already closed, skip error checking")
				return
			}

			// Handle user closing browser and terminating WebSocket connection
			// Handles "userleave" messages
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
					dataToSend := types.Message{
						Type:       WsUserLeave,
						Username:   currentName,
						TotalUsers: len(roomManager.Rooms[currentRoom]),
					}
					roomManager.SendMessageToAll(currentRoom, dataToSend)
				}
			} else {
				fmt.Println("Error reading message:", err)
			}
			break
		}

		// Update rate limiter here to also limit invalid messages getting received
		if !rateLimiter.AllowMessage() {
			fmt.Println("Rate limit exceeded, closing connection")

			if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
				dataToSend := types.Message{
					Type:       WsUserLeave,
					Username:   currentName,
					TotalUsers: len(roomManager.Rooms[currentRoom]),
				}
				roomManager.SendMessageToAll(currentRoom, dataToSend)
			}

			break
		}

		var msg types.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("Error unmarshalling message:", err)

			if usersLeft := roomManager.RemoveConnection(currentRoom, conn); usersLeft {
				dataToSend := types.Message{
					Type:       WsUserLeave,
					Username:   currentName,
					TotalUsers: len(roomManager.Rooms[currentRoom]),
				}
				roomManager.SendMessageToAll(currentRoom, dataToSend)
			}

			break
		}

		// Checks the type of message the user sent to the server
		switch msg.Type {
		case WsAuth:
			// Don't let users spam auth messages
			if authenticated {
				return
			}

			token := msg.Message
			if !auth.JWTTokenValid(token) {
				return
			}

			authenticated = true

			// Don't count the auth message towards the rate limiter
			rateLimiter.Reset()

			// Stop auth timeout
			close(authTimeoutChannel)

		case WsJoin:
			if !authenticated {
				fmt.Println("Not authenticated, can't join")
				return
			}

			// Don't let users spam join messages
			if hasJoined {
				return
			}

			currentName = msg.Username
			currentRoom = msg.Room
			if len(currentName) < 3 || len(currentRoom) < 3 {
				fmt.Println("Name and room have to be length >= 3")
				return
			}

			fmt.Printf("JOIN: '%s' room '%s'\n", currentName, currentRoom)

			roomManager.AddConnectionToRoom(currentRoom, conn)

			dataToSend := types.Message{
				Type:       WsJoin,
				Username:   currentName,
				TotalUsers: len(roomManager.Rooms[currentRoom]),
			}
			roomManager.SendMessageToAll(currentRoom, dataToSend)
			hasJoined = true

			// Don't count the join message towards rate limiter
			rateLimiter.Reset()

			// Stop join timeout
			close(joinTimeoutChannel)

		case WsUserLeave:
			fmt.Println("USERLEAVE:", currentName)
			break

		case WsMessage:
			if !authenticated || !hasJoined {
				fmt.Println("Can't send messages... user has not joined or been auth'd")
				return
			}

			fmt.Println("MESSAGE")

			if len(msg.Message) == 0 {
				fmt.Println("No message")
				break
			}

			fmt.Printf("\t%s: %s > %s\n", currentRoom, currentName, msg.Message)

			dataToSend := types.Message{
				Type:     WsMessage,
				Username: currentName,
				Message:  msg.Message,
			}
			roomManager.SendMessageToAllExceptSelf(conn, currentRoom, dataToSend)

		default:
			fmt.Printf("Unknown message type (%#v), closing connection\n", msg)
			return
		}
	}

	fmt.Println("End of WebSocket session for", currentName)
}

// The basic HTTP connection, not WebSocket yet
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, connErr := upgrader.Upgrade(w, r, nil)

	if connErr != nil {
		fmt.Println("Error upgrading connection:", connErr)
		return
	}

	go handleConnection(conn)
}
