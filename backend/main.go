package main

import (
	"ilovmath/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Load all Go HTML templates from the templates/ directory.
	r.LoadHTMLGlob("templates/*")

	// Serve static files (CSS, compiled JS, images, etc.).
	r.Static("/static", "./static")

	// Page routes (rendered by Go templates).
	r.GET("/", handlers.IndexPage)

	// REST API routes.
	api := r.Group("/api")
	{
		api.GET("/list", handlers.GetList)
	}

	log.Println("Server listening on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
