package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gheorghe_nc_website/internal/config"
)

const graphAPIBase = "https://graph.facebook.com/v23.0"
const templateLanguage = "en_US"

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("whatsapp api error %d: %s", e.Code, e.Message)
}

type WhatsAppService struct {
	cfg  *config.Config
	http *http.Client
}

func NewWhatsAppService(cfg *config.Config) *WhatsAppService {
	return &WhatsAppService{
		cfg:  cfg,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *WhatsAppService) Configured() bool {
	return s.cfg.WhatsAppToken != "" && s.cfg.WhatsAppPhoneNumberID != "" && s.cfg.OwnerPhoneNumber != ""
}

func (s *WhatsAppService) NotifyOwner(contactID int64, name, message string) (string, error) {
	if s.cfg.WhatsAppTemplateName != "" {
		return s.SendTemplate(s.cfg.OwnerPhoneNumber, s.cfg.WhatsAppTemplateName, fmt.Sprintf("#%d", contactID), name, message)
	}
	body := fmt.Sprintf(
		"New website enquiry #%d from %s:\n\n%s\n\nSwipe-reply to this message to respond.",
		contactID, name, message,
	)
	return s.SendText(s.cfg.OwnerPhoneNumber, body)
}

func (s *WhatsAppService) SendText(to, body string) (string, error) {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": body},
	}
	return s.send(payload)
}

func (s *WhatsAppService) SendTemplate(to, templateName string, params ...string) (string, error) {
	components := []any{}
	if len(params) > 0 {
		parameters := make([]any, 0, len(params))
		for _, p := range params {
			parameters = append(parameters, map[string]string{"type": "text", "text": p})
		}
		components = append(components, map[string]any{
			"type":       "body",
			"parameters": parameters,
		})
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "template",
		"template": map[string]any{
			"name":       templateName,
			"language":   map[string]string{"code": templateLanguage},
			"components": components,
		},
	}
	return s.send(payload)
}

func (s *WhatsAppService) send(payload any) (string, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", graphAPIBase, s.cfg.WhatsAppPhoneNumberID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("build whatsapp request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.WhatsAppToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call whatsapp api: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read whatsapp response: %w", err)
	}

	if res.StatusCode >= 300 {
		var errRes struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errRes) == nil && errRes.Error.Message != "" {
			return "", &APIError{Code: errRes.Error.Code, Message: errRes.Error.Message}
		}
		return "", fmt.Errorf("whatsapp api status %d: %s", res.StatusCode, string(body))
	}

	var okRes struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &okRes); err != nil {
		return "", fmt.Errorf("decode whatsapp response: %w", err)
	}
	if len(okRes.Messages) == 0 || okRes.Messages[0].ID == "" {
		return "", fmt.Errorf("whatsapp api response missing message id: %s", string(body))
	}
	return okRes.Messages[0].ID, nil
}

type InboundMessage struct {
	From      string `json:"from"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Text      *struct {
		Body string `json:"body"`
	} `json:"text"`
	Context *struct {
		MessageID string `json:"message_id"`
	} `json:"context"`
}

type WebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []InboundMessage `json:"messages"`
				Statuses []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"statuses"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}
