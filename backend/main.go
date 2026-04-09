package main

import (
	"ilovmath/config"
	"ilovmath/handlers"
	"ilovmath/math"
	"log"
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

	// REST API routes.
	api := r.Group("/api")
	{
		api.GET("/list", handlers.GetList)
		api.POST("/question/start", math.StartQuestion)
		api.POST("/question/next", math.NextQuestion)
	}

	log.Println("Server listening on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
