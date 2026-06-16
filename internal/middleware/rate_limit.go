package middleware

import (
	"net/http"
	"sync"
	"time"

	"zerobin-be/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > time.Hour {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func RateLimit(requestsPerSecond float64, burstSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := clients[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize)
			v = &client{
				limiter:  limiter,
				lastSeen: time.Now(),
			}
			clients[ip] = v
		} else {
			v.lastSeen = time.Now()
		}
		mu.Unlock()

		if !v.limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Too many requests. Please try again later.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
