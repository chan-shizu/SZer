package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "user_id"

type getSessionResponse struct {
	User *struct {
		ID string `json:"id"`
		} `json:"user"`
	}
	
	func frontendBaseURL() string {
	// In docker-compose, backend can reach frontend via http://frontend:3000
	if v := os.Getenv("BETTER_AUTH_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	if v := os.Getenv("FRONTEND_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://frontend:3000"
}

// FrontendBaseURL is used for building redirect URLs.
func FrontendBaseURL() string {
	return frontendBaseURL()
}

func RequireAuth() gin.HandlerFunc {
	client := &http.Client{Timeout: 5 * time.Second}
	log.Printf("[auth] RequireAuth middleware initialized")

	return func(c *gin.Context) {
		cookie := c.GetHeader("Cookie")
		if strings.TrimSpace(cookie) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, frontendBaseURL()+"/api/auth/get-session", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to build auth request"})
			return
		}
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Accept", "application/json")

		res, err := client.Do(req)
		if err != nil {
			// Fail closed: treat verification failure as unauthenticated.
			log.Printf("[auth] failed to verify session: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// better-auth returns `null` when not authenticated.
		if strings.TrimSpace(string(body)) == "" || strings.TrimSpace(string(body)) == "null" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var parsed getSessionResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if parsed.User == nil || strings.TrimSpace(parsed.User.ID) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(userIDContextKey, parsed.User.ID)
		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (string, error) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return "", errors.New("user_id not found")
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", errors.New("user_id invalid")
	}
	return s, nil
}

// 任意認証: クッキーがあればuserIDをContextにセット、なければ何もしない
func OptionalAuth() gin.HandlerFunc {
	client := &http.Client{Timeout: 5 * time.Second}
	return func(c *gin.Context) {
		cookie := c.GetHeader("Cookie")
		if strings.TrimSpace(cookie) == "" {
			c.Next()
			return
		}
		req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, frontendBaseURL()+"/api/auth/get-session", nil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Accept", "application/json")
		res, err := client.Do(req)
		if err != nil || res.StatusCode != http.StatusOK {
			c.Next()
			return
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		var parsed getSessionResponse
		if err := json.Unmarshal(body, &parsed); err == nil && parsed.User != nil && parsed.User.ID != "" {
			c.Set(userIDContextKey, parsed.User.ID)
		}
		c.Next()
	}
}