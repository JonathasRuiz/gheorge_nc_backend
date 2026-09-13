package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"gheorghe_nc_website/internal/config"
	"gheorghe_nc_website/internal/handlers"
	"gheorghe_nc_website/internal/middleware"
	"gheorghe_nc_website/internal/services"
	"gheorghe_nc_website/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer st.Close()

	emailService := services.NewEmailService(cfg)
	openaiService := services.NewOpenAIService(cfg)
	whatsappService := services.NewWhatsAppService(cfg)

	contactHandler := handlers.NewContactHandler(emailService, whatsappService, st)
	chatHandler := handlers.NewChatHandler(openaiService)
	webhookHandler := handlers.NewWebhookHandler(cfg, st, emailService)

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api", middleware.RateLimit(30, time.Minute))
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		api.POST("/contact", contactHandler.Handle)
		api.POST("/chat", chatHandler.Handle)
	}

	r.GET("/webhook/whatsapp", webhookHandler.Verify)
	r.POST("/webhook/whatsapp", webhookHandler.Handle)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
