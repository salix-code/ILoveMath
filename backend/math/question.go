package math

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ilovmath/config"
	"ilovmath/session"
)

type startRequest struct {
	ID         int    `json:"id" binding:"required"`
	Difficulty int    `json:"difficulty" binding:"required,min=1,max=3"`
	Order      string `json:"order"`
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
	if req.Order == "sequential" {
		cfg.OrderMode = "sequential"
	} else {
		cfg.OrderMode = "random"
	}
	cfg.QuestionIndex = 0
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
	GUID          string `json:"guid"`
	TypeLabel     string `json:"type_label"`
	Content       string `json:"content"`
	Score         int    `json:"score"`
	Total         int    `json:"total"`
	QuestionCount int    `json:"question_count"`
	Correct       *bool  `json:"correct,omitempty"` // nil on first question
}

var difficultyLabel = map[int]string{1: "低", 2: "中", 3: "高"}

// textGenerators maps function names (used in Answer.Text as "{FuncName}" or
// "{FuncName(arg1,arg2)}") to generator functions that return (question content,
// answer value). The string slice carries any arguments parsed from the call syntax.
var textGenerators = map[string]func([]string) (string, string){
	"GenerateExpression": func(args []string) (string, string) {
		println(strings.Join(args, ", "))
		eq, answer := GenerateExpression(args...)
		//println(eq.Equation, eq.AValue)
		return eq, strconv.Itoa(answer)
	},
}

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
	question, exists := config.ProblemTypes[ses.ProblemID]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown problem type"})
		return
	}

	var candidates []config.ProblemItem
	for _, item := range question.Items {
		if item.Difficulty == ses.Difficulty {
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 0 {
		candidates = question.Items
	}
	if len(candidates) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no questions available"})
		return
	}
	var item config.ProblemItem
	if ses.OrderMode == "sequential" {
		item = candidates[ses.QuestionIndex%len(candidates)]
		ses.QuestionIndex++
	} else {
		item = candidates[rand.Intn(len(candidates))]
	}

	// 3. Evaluate Input expressions → concrete integer map.
	resolved, err := resolveInput(item.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("input evaluation: %v", err)})
		return
	}

	// Substitute {key} placeholders in the question text.
	content := substituteQuestion(item.Question, resolved)

	// 4. Select a single answer item and attach its text.
	if len(item.Answer) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no answers defined for this question"})
		return
	}
	selectedAnswer := item.Answer[rand.Intn(len(item.Answer))]
	var generatedAnswer string
	if selectedAnswer.Text != "" {
		text := selectedAnswer.Text
		if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
			funcCall := text[1 : len(text)-1]
			funcName, args := parseFuncCall(funcCall)
			if gen, ok := textGenerators[funcName]; ok {
				genContent, genAnswer := gen(args)
				content = fmt.Sprintf("\n%s \n%s", content, genContent)
				generatedAnswer = genAnswer
			}
		} else {
			answerText := substituteQuestion(selectedAnswer.Text, resolved)
			content = fmt.Sprintf("%s \n%s？", content, answerText)
		}
	}

	// Compute result for the selected answer.
	var answerValue string
	if generatedAnswer != "" {
		answerValue = generatedAnswer
	} else if v, err := evalExpr(selectedAnswer.Value, resolved); err == nil {
		answerValue = strconv.Itoa(v)
	} else {
		answerValue = substituteQuestion(selectedAnswer.Value, resolved)
	}

	// 5. Record answer + new GUID in session for next-request validation.
	guid := uuid.NewString()
	ses.CurrentGUID = guid
	ses.CurrentAnswer = answerValue

	c.JSON(http.StatusOK, nextResponse{
		GUID:          guid,
		TypeLabel:     fmt.Sprintf("%s — 难度：%s", question.Title, difficultyLabel[ses.Difficulty]),
		Content:       content,
		Score:         ses.Score,
		Total:         ses.Total,
		QuestionCount: len(question.Items),
		Correct:       correct,
	})
}

// parseFuncCall parses a string like "FuncName(a, b)" into the function name
// and a slice of trimmed string arguments. If there are no parentheses the
// whole string is treated as the function name with no arguments.
func parseFuncCall(s string) (string, []string) {
	idx := strings.Index(s, "(")
	if idx == -1 {
		return s, nil
	}
	funcName := s[:idx]
	argStr := strings.TrimSuffix(s[idx+1:], ")")
	if argStr == "" {
		return funcName, nil
	}
	parts := strings.Split(argStr, ",")
	args := make([]string, len(parts))
	for i, p := range parts {
		args[i] = strings.TrimSpace(p)
	}
	return funcName, args
}

func ListQuestions(c *gin.Context) {

}
