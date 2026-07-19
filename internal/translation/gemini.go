package translation

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
)

const inboundSystemPrompt = `You are translating Path of Exile chat messages.

Translate the message into natural Japanese.

Rules:
- Understand Path of Exile terminology and slang.
- Keep item names, skill names, and currency names in English when appropriate.
- Keep common PoE abbreviations (div, ex, chaos, etc.) when appropriate.
- Do not explain the translation.
- Return only the Japanese translation.`

const outboundSystemPrompt = `You are helping a Japanese Path of Exile player reply to an English-speaking player.

Convert the Japanese message into natural, concise English suitable for Path of Exile chat.

Rules:
- Preserve the user's intended meaning.
- Use natural PoE chat language.
- Keep it concise as it will be typed into game chat.
- Do not add information the user did not provide.
- Do not explain the translation.
- Return only the English message.`

type GeminiTranslator struct {
	client *genai.Client
	model  string
}

func NewGemini(ctx context.Context, apiKey string) (*GeminiTranslator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiTranslator{
		client: client,
		model:  "gemini-2.0-flash",
	}, nil
}

func (g *GeminiTranslator) Translate(ctx context.Context, req Request) (string, error) {
	systemPrompt := inboundSystemPrompt
	if req.Direction == Outbound {
		systemPrompt = outboundSystemPrompt
	}

	userMessage := req.Message
	if req.Direction == Outbound && len(req.Context) > 0 {
		userMessage = buildOutboundPrompt(req.Message, req.Context)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		Temperature:       genai.Ptr(float32(0.3)),
		MaxOutputTokens:   256,
	}

	result, err := g.client.Models.GenerateContent(ctx, g.model, []*genai.Content{genai.NewContentFromText(userMessage, genai.RoleUser)}, config)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}

	return extractText(result), nil
}

func buildOutboundPrompt(message string, context []chat.Message) string {
	var b strings.Builder
	b.WriteString("Context:\n")
	for _, m := range context {
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Player, m.Body))
	}
	b.WriteString(fmt.Sprintf("\nUser:\n%s", message))
	return b.String()
}

func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return ""
	}
	var texts []string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.TrimSpace(strings.Join(texts, ""))
}
