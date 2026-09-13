package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gheorghe_nc_website/internal/config"
	"gheorghe_nc_website/internal/services"
	"gheorghe_nc_website/internal/store"
)

type WebhookHandler struct {
	cfg          *config.Config
	store        *store.Store
	emailService *services.EmailService
	ownerDigits  string
}

func NewWebhookHandler(cfg *config.Config, st *store.Store, emailService *services.EmailService) *WebhookHandler {
	return &WebhookHandler{
		cfg:          cfg,
		store:        st,
		emailService: emailService,
		ownerDigits:  normalizeDigits(cfg.OwnerPhoneNumber),
	}
}

func (h *WebhookHandler) Verify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && h.cfg.WhatsAppVerifyToken != "" && token == h.cfg.WhatsAppVerifyToken {
		c.Data(http.StatusOK, "text/plain", []byte(challenge))
		return
	}

	log.Printf("[webhook] verification rejected: mode=%q", mode)
	c.Status(http.StatusForbidden)
}

func (h *WebhookHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[webhook] read body failed: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	if h.cfg.WhatsAppAppSecret == "" {
		log.Println("[webhook] WARNING: WHATSAPP_APP_SECRET not set; skipping signature verification")
	} else if !verifySignature(h.cfg.WhatsAppAppSecret, body, c.GetHeader("X-Hub-Signature-256")) {
		log.Printf("[webhook] invalid X-Hub-Signature-256")
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload services.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[webhook] malformed payload: %v", err)
		c.Status(http.StatusOK)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				h.processMessage(msg, body)
			}
		}
	}

	c.Status(http.StatusOK)
}

func (h *WebhookHandler) processMessage(msg services.InboundMessage, raw []byte) {
	isNew, err := h.store.InsertWebhookEvent(msg.ID, string(raw))
	if err != nil {
		log.Printf("[webhook] store event %s failed: %v", msg.ID, err)
		return
	}
	if !isNew {
		log.Printf("[webhook] duplicate delivery ignored: %s", msg.ID)
		return
	}

	from := normalizeDigits(msg.From)
	if from != h.ownerDigits || h.ownerDigits == "" {
		log.Printf("[webhook] ignoring message from non-owner sender: %s", msg.From)
		return
	}

	if msg.Type != "text" || msg.Text == nil {
		log.Printf("[webhook] ignoring non-text message from owner: %s (type=%s)", msg.ID, msg.Type)
		return
	}

	contact, method := h.correlate(msg)
	if contact == nil {
		log.Printf("[webhook] no contacts stored; ignoring owner message %s", msg.ID)
		return
	}

	replyBody := strings.TrimSpace(stripTagPrefix(msg.Text.Body))
	if replyBody == "" {
		replyBody = msg.Text.Body
	}

	if err := h.emailService.SendReplyEmail(contact.Email, replyBody); err != nil {
		log.Printf("[webhook] forward reply to %s failed: %v", contact.Email, err)
	} else {
		log.Printf("[webhook] reply %s forwarded to %s (contact #%d, method=%s)", msg.ID, contact.Email, contact.ID, method)
	}

	if err := h.store.InsertReply(contact.ID, msg.ID, replyBody, method); err != nil {
		log.Printf("[webhook] record reply %s failed: %v", msg.ID, err)
	}
}

func (h *WebhookHandler) correlate(msg services.InboundMessage) (*store.Contact, string) {
	if msg.Context != nil && msg.Context.MessageID != "" {
		contact, err := h.store.GetContactByNotifyMsgID(msg.Context.MessageID)
		if err != nil {
			log.Printf("[webhook] lookup by notify_msg_id failed: %v", err)
		}
		if contact != nil {
			return contact, "quote"
		}
	}

	if id, ok := parseTagPrefix(msg.Text.Body); ok {
		contact, err := h.store.GetContactByID(id)
		if err != nil {
			log.Printf("[webhook] lookup by id tag failed: %v", err)
		}
		if contact != nil {
			return contact, "tag"
		}
	}

	contact, err := h.store.LatestContact()
	if err != nil {
		log.Printf("[webhook] latest contact lookup failed: %v", err)
	}
	if contact != nil {
		return contact, "fallback"
	}
	return nil, ""
}

func verifySignature(secret string, body []byte, header string) bool {
	if !strings.HasPrefix(header, "sha256=") {
		return false
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(sig, mac.Sum(nil))
}

func normalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseTagPrefix(body string) (int64, bool) {
	if !strings.HasPrefix(body, "#") {
		return 0, false
	}
	i := 1
	for i < len(body) && body[i] >= '0' && body[i] <= '9' {
		i++
	}
	if i == 1 || (i < len(body) && body[i] >= '0' && body[i] <= '9') {
		return 0, false
	}
	id, err := strconv.ParseInt(body[1:i], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func stripTagPrefix(body string) string {
	rest := strings.TrimPrefix(body, "#")
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 {
		return body
	}
	return strings.TrimLeft(rest[i:], " \t")
}
