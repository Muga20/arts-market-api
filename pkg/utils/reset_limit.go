package utils

import (
	"sync"
	"time"
)

type RateLimiter struct {
	// A map to store the email and the last request time
	emailRequests map[string]time.Time
	mu            sync.Mutex
	limitDuration time.Duration
}

var limiter = RateLimiter{
	emailRequests: make(map[string]time.Time),
	limitDuration: 1 * time.Minute, // Example: 1 minute cooldown
}

// IsRateLimited checks if the email has made a request within the allowed time window
func IsRateLimited(email string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Check if the email exists in the map and if it was recently used
	lastRequestTime, exists := limiter.emailRequests[email]
	if !exists {
		// First request or reset, allow it
		limiter.emailRequests[email] = time.Now()
		return false
	}

	// If the last request time is within the limit duration, rate limit the request
	if time.Since(lastRequestTime) < limiter.limitDuration {
		return true
	}

	// Update the last request time for the email
	limiter.emailRequests[email] = time.Now()
	return false
}
