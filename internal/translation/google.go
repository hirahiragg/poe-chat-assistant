package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GoogleTranslator struct {
	client *http.Client
}

func NewGoogle() *GoogleTranslator {
	return &GoogleTranslator{client: &http.Client{}}
}

func (g *GoogleTranslator) Translate(ctx context.Context, req Request) (string, error) {
	sl, tl := "en", "ja"
	if req.Direction == Outbound {
		sl, tl = "ja", "en"
	}

	u := "https://translate.googleapis.com/translate_a/single?" + url.Values{
		"client": {"gtx"},
		"sl":     {sl},
		"tl":     {tl},
		"dt":     {"t"},
		"q":      {req.Message},
	}.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("google translate request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("google translate read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google translate status %d: %s", resp.StatusCode, body)
	}

	return parseGoogleResponse(body)
}

func parseGoogleResponse(body []byte) (string, error) {
	var data []any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("google translate parse: %w", err)
	}

	sentences, ok := data[0].([]any)
	if !ok {
		return "", fmt.Errorf("google translate: unexpected response format")
	}

	var result string
	for _, s := range sentences {
		parts, ok := s.([]any)
		if !ok || len(parts) == 0 {
			continue
		}
		if text, ok := parts[0].(string); ok {
			result += text
		}
	}
	return result, nil
}
