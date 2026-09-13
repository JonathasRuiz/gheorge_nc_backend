package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gheorghe_nc_website/internal/services"
	"gheorghe_nc_website/internal/store"
)

type ContactHandler struct {
	emailService *services.EmailService
	whatsapp     *services.WhatsAppService
	store        *store.Store
}

func NewContactHandler(emailService *services.EmailService, whatsapp *services.WhatsAppService, st *store.Store) *ContactHandler {
	return &ContactHandler{emailService: emailService, whatsapp: whatsapp, store: st}
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

	id, err := h.store.CreateContact(req.Name, req.Email, req.Phone, req.Message)
	if err != nil {
		log.Printf("[contact] save submission failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit contact form"})
		return
	}

	if err := h.emailService.SendContactEmail(req.Name, req.Email, req.Phone, req.Message); err != nil {
		log.Printf("[contact] #%d forward email failed: %v", id, err)
	}

	h.notifyOwnerWhatsApp(id, req.Name, req.Message)

	c.JSON(http.StatusOK, gin.H{"message": "Contact form submitted successfully"})
}

func (h *ContactHandler) notifyOwnerWhatsApp(id int64, name, message string) {
	if !h.whatsapp.Configured() {
		log.Printf("[contact] #%d WhatsApp notification skipped: WHATSAPP_TOKEN/WHATSAPP_PHONE_NUMBER_ID/OWNER_PHONE_NUMBER not fully set", id)
		h.recordWhatsAppResult(id, "", "skipped_no_config")
		return
	}

	msgID, err := h.whatsapp.NotifyOwner(id, name, message)
	if err != nil {
		log.Printf("[contact] #%d WhatsApp notification failed: %v", id, err)
		status := "failed"
		var apiErr *services.APIError
		if errors.As(err, &apiErr) {
			status = fmt.Sprintf("failed:%d", apiErr.Code)
		}
		h.recordWhatsAppResult(id, "", status)
		return
	}

	log.Printf("[contact] #%d WhatsApp notification sent to owner (wamid=%s)", id, msgID)
	h.recordWhatsAppResult(id, msgID, "sent")
}

func (h *ContactHandler) recordWhatsAppResult(id int64, msgID, status string) {
	if err := h.store.UpdateContactWhatsAppResult(id, msgID, status); err != nil {
		log.Printf("[contact] #%d record WhatsApp result failed: %v", id, err)
	}
}
