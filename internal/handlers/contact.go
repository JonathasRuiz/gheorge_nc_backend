package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gheorghe_nc_website/internal/services"
)

type ContactHandler struct {
	emailService *services.EmailService
}

func NewContactHandler(emailService *services.EmailService) *ContactHandler {
	return &ContactHandler{emailService: emailService}
}

type ContactRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Phone   string `json:"phone" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func (h *ContactHandler) Handle(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.emailService.SendContactEmail(req.Name, req.Email, req.Phone, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact form submitted successfully"})
}
