package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type     string `json:"type"`
	Message  string `json:"message,omitempty"`
	Username string `json:"username,omitempty"`
	Room     string `json:"room,omitempty"`
}

var room_manager *RoomManager = NewRoomManager()

func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	var currentName string
	var currentRoom string

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			// Handle user closing browser and terminating WebSocket connection
			// Handles "userleave" messages
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				fmt.Println("User has left/closed the room")
				if _, exists := room_manager.Rooms[currentRoom]; exists {
					if users_left := room_manager.RemoveConnection(currentRoom, conn); users_left {
						fmt.Println("\tBroadcasting USERLEAVE to all...")
						data_to_send := map[string]interface{}{
							"type":       "userleave",
							"user":       currentName,
							"totalUsers": len(room_manager.Rooms[currentRoom]),
						}

						json_data, err := json.Marshal(data_to_send)
						if err != nil {
							fmt.Println("JSON Marhsall error:", err)
							return
						}
						for otherConn := range room_manager.Rooms[currentRoom] {
							otherConn.WriteMessage(websocket.TextMessage, json_data)
						}
					}
				} else {
					fmt.Printf("Room %s not exist", currentRoom)
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
		switch msg.Type {
		case "join":
			currentName = msg.Username
			currentRoom = msg.Room
			fmt.Println("JOIN:", currentName, currentRoom)

			if currentName == "" || currentRoom == "" {
				fmt.Println("Empty values")
				continue
			}

			room_manager.AddConnection(currentRoom, conn)

			roomConns := room_manager.Rooms[currentRoom]

			fmt.Printf("\t'%s' has joined room '%s'\n", currentName, currentRoom)

			data_to_send := map[string]interface{}{
				"type":       "newuser",
				"user":       currentName,
				"totalUsers": len(room_manager.Rooms[currentRoom]),
			}

			json_data, err := json.Marshal(data_to_send)
			if err != nil {
				fmt.Println("JSON Marhsall error:", err)
				return
			}
			for otherConn := range roomConns {
				otherConn.WriteMessage(websocket.TextMessage, json_data)
			}
		case "userleave":
			fmt.Println("USERLEAVE:", currentName)
			break

		case "message":
			fmt.Println("MESSAGE")
			fmt.Printf("\t%s: %s > %s\n", currentRoom, currentName, msg.Message)

			roomConns := room_manager.Rooms[currentRoom]

			data_to_send := map[string]interface{}{
				"type":    "message",
				"user":    currentName,
				"message": msg.Message,
			}

			json_data, err := json.Marshal(data_to_send)
			if err != nil {
				fmt.Println("JSON Marhsall error:", err)
				return
			}
			for otherConn := range roomConns {
				if otherConn == conn {
					continue
				}
				otherConn.WriteMessage(websocket.TextMessage, json_data)
			}
		}
	}

	fmt.Println("Client connected")
	fmt.Println("End of WebSocket session")

	// todo: clean up leftover users and rooms
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token_str := r.URL.Query().Get("token")
	if token_str == "" {
		fmt.Println("No Authorization token")
		return
	}

	sb_secret := os.Getenv("SUPABASE_SECRET")
	if sb_secret == "" {
		fmt.Println("No supabase secret")
		return
	}

	tok, jwt_err := jwt.Parse(token_str, func(token *jwt.Token) (interface{}, error) {
		return []byte(sb_secret), nil
	})

	if jwt_err != nil {
		fmt.Println("Error trying to parse JWT:", jwt_err)
		return
	}

	if !tok.Valid {
		fmt.Println("Invalid token")
		return
	}

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

	http.HandleFunc("/", handleWebSocket)

	fmt.Println("WebSocket server running on ws://localhost:8080/")
	err := http.ListenAndServe(":8080", nil)
	if goDotEnv_err != nil {
		fmt.Println("Error starting server:", err)
	}
}
