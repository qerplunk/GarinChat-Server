package ratelimiter

import (
	"time"
)

// The frontend will hard limit users sending messages to once every 2.5 seconds
// But due to network timings the server limits to 1 message/second to avoid kicking good users out
const (
	maxMessagesPerSecond = 1
	slidingWindowSeconds = 1
)

type RateLimiter struct {
	// Track the number of messages received in the current time window
	MessageCount int
	// Store the timestamp of the last time the counter 'messageCount' was reset
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
	// Reset rate limiter
	rl.MessageCount = 0
	rl.LastReset = time.Now()
}
