package ratelimiter

import (
	"time"
)

// Default messages per time window the server expects from a frontend user
// Configure based on your needs
const (
	maxMessagesPerSecond = 1
	slidingWindowSeconds = 1
)

type RateLimiter struct {
	// Track the number of messages received in the current time window
	MessageCount int
	// Defines the start of the current time window
	LastReset time.Time
}

// Returns if the user is under the rate limit and allowed to send messages to the server
func (rl *RateLimiter) AllowMessage() bool {
	now := time.Now()

	// Reset rate limiter every slidingWindowSeconds
	if now.Sub(rl.LastReset) > slidingWindowSeconds*time.Second {
		rl.MessageCount = 0
		rl.LastReset = now
	}

	if rl.MessageCount >= maxMessagesPerSecond {
		return false
	}

	// User did not go over rate limit
	rl.MessageCount += 1
	return true
}

// Reset the rate limiter.
func (rl *RateLimiter) Reset() {
	rl.MessageCount = 0
	rl.LastReset = time.Now()
}
