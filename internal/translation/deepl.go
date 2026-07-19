package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type DeepLTranslator struct {
	client *http.Client
	apiKey string
	apiURL string
}

func NewDeepL(apiKey string) *DeepLTranslator {
	apiURL := "https://api-free.deepl.com/v2/translate"
	if !strings.Contains(apiKey, ":fx") {
		apiURL = "https://api.deepl.com/v2/translate"
	}
	return &DeepLTranslator{
		client: &http.Client{},
		apiKey: apiKey,
		apiURL: apiURL,
	}
}

func (d *DeepLTranslator) Translate(ctx context.Context, req Request) (string, error) {
	sl, tl := "EN", "JA"
	if req.Direction == Outbound {
		sl, tl = "JA", "EN"
	}

	form := url.Values{
		"text":        {req.Message},
		"source_lang": {sl},
		"target_lang": {tl},
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "DeepL-Auth-Key "+d.apiKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("deepl request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("deepl read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepl status %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("deepl parse: %w", err)
	}
	if len(result.Translations) == 0 {
		return "", fmt.Errorf("deepl: empty response")
	}
	return result.Translations[0].Text, nil
}
