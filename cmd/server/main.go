package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"gheorghe_nc_website/internal/config"
	"gheorghe_nc_website/internal/handlers"
	"gheorghe_nc_website/internal/middleware"
	"gheorghe_nc_website/internal/services"
)

func main() {
	cfg := config.Load()

	emailService := services.NewEmailService(cfg)
	openaiService := services.NewOpenAIService(cfg)

	contactHandler := handlers.NewContactHandler(emailService)
	chatHandler := handlers.NewChatHandler(openaiService)

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimit(30, time.Minute))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/contact", contactHandler.Handle)
	r.POST("/api/chat", chatHandler.Handle)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
