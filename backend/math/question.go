package math

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ilovmath/config"
	"ilovmath/session"
)

type startRequest struct {
	ID         int `json:"id" binding:"required"`
	Difficulty int `json:"difficulty" binding:"required,min=1,max=3"`
}

// StartQuestion handles POST /api/question/start.
func StartQuestion(c *gin.Context) {
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and difficulty (1-3) are required"})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session"})
		return
	}

	cfg := session.GetOrCreate(sessionID)
	cfg.ProblemID = req.ID
	cfg.Difficulty = req.Difficulty
	cfg.Score = 0
	cfg.Total = 0
	cfg.CurrentGUID = ""
	cfg.CurrentAnswer = ""

	c.JSON(http.StatusOK, gin.H{"redirect": "/question"})
}

type nextRequest struct {
	PrevGUID   string `json:"prev_guid"`
	PrevAnswer string `json:"prev_answer"`
}

type nextResponse struct {
	GUID      string `json:"guid"`
	TypeLabel string `json:"type_label"`
	Content   string `json:"content"`
	Score     int    `json:"score"`
	Total     int    `json:"total"`
	Correct   *bool  `json:"correct,omitempty"` // nil on first question
}

var difficultyLabel = map[int]string{1: "低", 2: "中", 3: "高"}

// NextQuestion handles POST /api/question/next.
func NextQuestion(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session"})
		return
	}

	ses, ok := session.Get(sessionID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
		return
	}

	var req nextRequest
	_ = c.ShouldBindJSON(&req) // body is optional on the first call

	// 1. Validate previous answer if one was submitted.
	var correct *bool
	if req.PrevGUID != "" && req.PrevGUID == ses.CurrentGUID && ses.CurrentAnswer != "" {
		ses.Total++
		isCorrect := strings.TrimSpace(req.PrevAnswer) == ses.CurrentAnswer
		correct = &isCorrect
		if isCorrect {
			ses.Score++
		}
	}

	// 2. Collect candidates matching the session difficulty; fall back to any.
	prob, exists := config.ProblemTypes[ses.ProblemID]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown problem type"})
		return
	}

	var candidates []config.ProblemItem
	for _, item := range prob.Items {
		if item.Difficulty == ses.Difficulty {
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 0 {
		candidates = prob.Items
	}
	if len(candidates) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no questions available"})
		return
	}
	item := candidates[rand.Intn(len(candidates))]

	// 3. Evaluate Input expressions → concrete integer map.
	resolved, err := resolveInput(item.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("input evaluation: %v", err)})
		return
	}

	// Substitute {key} placeholders in the question text.
	content := substituteQuestion(item.Question, resolved)

	// 4. Compute the expected answer string (arithmetic expr or mixed text).
	answerStr := evalAnswer(item.Answer, resolved)

	// 5. Record answer + new GUID in session for next-request validation.
	guid := uuid.NewString()
	ses.CurrentGUID = guid
	ses.CurrentAnswer = answerStr

	c.JSON(http.StatusOK, nextResponse{
		GUID:      guid,
		TypeLabel: fmt.Sprintf("%s — 难度：%s", prob.Title, difficultyLabel[ses.Difficulty]),
		Content:   content,
		Score:     ses.Score,
		Total:     ses.Total,
		Correct:   correct,
	})
}
