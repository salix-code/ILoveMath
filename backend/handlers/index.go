package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// ProblemType represents a category of math problems.
type ProblemType struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// SessionConfig holds per-session configuration.
type SessionConfig struct {
	ID string
}

var (
	sessionStore sync.Map

	problemTypes = []ProblemType{
		{ID: 1, Title: "和差问题"},
		{ID: 2, Title: "倍数问题"},
		{ID: 3, Title: "行程问题"},
		{ID: 4, Title: "鸡兔同笼"},
		{ID: 5, Title: "植树问题"},
	}
)

func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// IndexPage handles GET / — renders the main page via Go template.
func IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

// GetList handles GET /api/list.
// Creates or validates a session and returns all problem types.
func GetList(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")

	if sessionID == "" {
		id, err := generateSessionID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
			return
		}
		sessionID = id
		sessionStore.Store(sessionID, &SessionConfig{ID: sessionID})
	} else {
		if _, ok := sessionStore.Load(sessionID); !ok {
			sessionStore.Store(sessionID, &SessionConfig{ID: sessionID})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"items":      problemTypes,
	})
}
