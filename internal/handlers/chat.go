package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"

	"gheorghe_nc_website/internal/services"
)

type ChatHandler struct {
	openaiService *services.OpenAIService
}

func NewChatHandler(openaiService *services.OpenAIService) *ChatHandler {
	return &ChatHandler{openaiService: openaiService}
}

type ChatRequest struct {
	Messages []openai.ChatCompletionMessage `json:"messages" binding:"required"`
}

func (h *ChatHandler) Handle(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	c.Writer.Flush()

	err := h.openaiService.StreamChat(ctx, req.Messages, c.Writer)
	if err != nil {
		fmt.Fprintf(c.Writer, "\n\n[ERROR] %s", err.Error())
		c.Writer.Flush()
	}
}
