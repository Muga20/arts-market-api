package middleware

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

var limiters = make(map[string]*rate.Limiter)
var mu sync.Mutex

// RateLimitMiddleware applies a rate limiter to requests based on their IP address
func RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		mu.Lock()
		defer mu.Unlock()

		// If the IP doesn't exist in the map, create a new limiter for that IP
		if _, exists := limiters[ip]; !exists {
			limiters[ip] = rate.NewLimiter(rate.Every(time.Minute), 20) // Allow 5 requests per minute
		}

		limiter := limiters[ip]

		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later",
			})
		}

		return c.Next()
	}
}
