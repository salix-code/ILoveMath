package main

import (
	"ilovmath/config"
	"ilovmath/handlers"
	"ilovmath/math"
	"ilovmath/session"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadAll("config"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("loaded %d problem type(s) from config/", len(config.ProblemTypes))

	r := gin.Default()

	// Disable browser cache for static files during development.
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Load all Go HTML templates from the templates/ directory.
	r.LoadHTMLGlob("templates/*")

	// Serve static files (CSS, compiled JS, images, etc.).
	r.Static("/static", "./static")

	// Page routes (rendered by Go templates).
	r.GET("/", handlers.IndexPage)
	r.GET("/question", handlers.QuestionPage)
	r.GET("/paper", handlers.PaperPage)

	// REST API routes.
	api := r.Group("/api")
	{
		api.GET("/list", handlers.GetList)
		api.POST("/question/start", func(c *gin.Context) {
			var req struct {
				ID         int    `json:"id"         binding:"required"`
				Difficulty int    `json:"difficulty" binding:"required,min=1,max=3"`
				Action     string `json:"action"`
			}
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

			redirect := "/question"
			if req.Action == "print" {
				redirect = "/paper"
			}
			c.JSON(http.StatusOK, gin.H{"redirect": redirect})
		})
		api.POST("/question/next", math.NextQuestion)

		api.GET("/question/list", math.ListQuestions)
	}

	log.Println("Server listening on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
