package services

import (
	"context"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"

	"gheorghe_nc_website/internal/config"
)

type OpenAIService struct {
	client *openai.Client
	model  string
}

func NewOpenAIService(cfg *config.Config) *OpenAIService {
	client := openai.NewClient(cfg.OpenAIAPIKey)
	return &OpenAIService{
		client: client,
		model:  cfg.OpenAIModel,
	}
}

func (s *OpenAIService) StreamChat(ctx context.Context, messages []openai.ChatCompletionMessage, writer io.Writer) error {
	systemPrompt := `You are a helpful assistant for Gheorghe NC, a professional tiling company.

Company Info:
- Name: Gheorghe NC
- Services: Floor tiling, wall tiling, bathroom tiling, kitchen tiling, outdoor tiling, tile repairs
- Experience: 15+ years in the tiling industry
- Location: London, UK
- Phone: 01234 567 890
- Email: info@gheorghenc.co.uk
- Opening Hours: Mon-Fri 8am-6pm, Sat 9am-4pm, Sun closed

Pricing Guidance:
- We offer free consultations and quotes
- Prices vary based on tile type, area size, and complexity
- We use premium materials and provide a warranty on our work

FAQs:
- We are fully insured with public liability insurance
- We handle both residential and commercial projects
- We can source tiles or work with customer-supplied materials
- Typical bathroom renovation takes 5-7 working days
- We offer underfloor heating installation

Be friendly, professional, and helpful. If someone asks about pricing, explain that we provide free quotes and encourage them to call or fill out the contact form. Keep responses concise.`

	allMessages := append(
		[]openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		},
		messages...,
	)

	stream, err := s.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    s.model,
			Messages: allMessages,
			Stream:   true,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			if chunk != "" {
				if _, err := writer.Write([]byte(chunk)); err != nil {
					return fmt.Errorf("write error: %w", err)
				}
			}
		}
	}

	return nil
}
