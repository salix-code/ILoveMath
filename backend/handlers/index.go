package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"ilovmath/config"
	"ilovmath/session"
)

// ProblemType is the list-item DTO returned to clients.
type ProblemType struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

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

// QuestionPage handles GET /question — renders the question page.
func QuestionPage(c *gin.Context) {
	c.HTML(http.StatusOK, "question.html", gin.H{})
}

func PaperPage(c *gin.Context) {
	c.HTML(http.StatusOK, "paper.html", gin.H{})
}

// GetList handles GET /api/list.
// Creates or validates a session and returns all problem types loaded from config.
func GetList(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")

	if sessionID == "" {
		id, err := generateSessionID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
			return
		}
		sessionID = id
	}

	session.GetOrCreate(sessionID)

	items := make([]ProblemType, 0, len(config.ProblemTypes))
	for _, cfg := range config.ProblemTypes {
		items = append(items, ProblemType{ID: cfg.ID, Title: cfg.Title})
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"items":      items,
	})
}
