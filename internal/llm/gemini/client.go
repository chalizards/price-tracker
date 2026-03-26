package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/chalizards/price-tracker/internal/llm"
	"github.com/chalizards/price-tracker/internal/sanitizer"
)

const apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent"

func ExtractPrice(ctx context.Context, apiKey string, html string, productName string) (*llm.PriceResult, error) {
	cleanHTML := sanitizer.SanitizeForLLM(html)
	prompt := buildPriceExtractionPrompt(productName, cleanHTML)

	reqBody := request{
		SystemInstruction: &content{Parts: []part{{Text: systemInstruction}}},
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("gemini api returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		log.Printf("[gemini] status=%d, headers=%v, body=%s", resp.StatusCode, resp.Header, string(body))
		return nil, fmt.Errorf("gemini api returned status %d", resp.StatusCode)
	}

	return parseResponse(resp)
}

func parseResponse(resp *http.Response) (*llm.PriceResult, error) {
	var geminiResp response
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("[gemini] raw response text: %s", text)

	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	log.Printf("[gemini] cleaned text: %s", text)

	var result llm.PriceResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse price from response: %w (raw: %s)", err, text)
	}

	if len(result.Prices) == 0 {
		return nil, fmt.Errorf("no prices returned from LLM")
	}

	for _, entry := range result.Prices {
		if entry.Price <= 0 {
			return nil, fmt.Errorf("invalid price from LLM: %.2f", entry.Price)
		}
		if entry.Currency == "" {
			return nil, fmt.Errorf("empty currency from LLM")
		}
		if entry.PaymentType != "pix" && entry.PaymentType != "credit" {
			return nil, fmt.Errorf("invalid payment_type from LLM: %s", entry.PaymentType)
		}
		log.Printf("[gemini] parsed: %.2f %s (%s)", entry.Price, entry.Currency, entry.PaymentType)
	}

	return &result, nil
}
