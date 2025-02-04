package types

// Message type used for unmarshalling WebSocket message into a variable
type Message struct {
	Type       string `json:"type"`
	Body       string `json:"body,omitempty"`
	Username   string `json:"username,omitempty"`
	Room       string `json:"room,omitempty"`
	TotalUsers int    `json:"totalUsers,omitempty"`
}
