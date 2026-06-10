package compliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type embedFnType func(ctx context.Context, text string) ([]float32, error)

type MistralEmbedder struct {
	embedFn embedFnType
}

func NewMistralEmbedder(_ context.Context, apiKey string) (*MistralEmbedder, error) {
	e := &MistralEmbedder{}
	e.embedFn = func(ctx context.Context, text string) ([]float32, error) {
		return embedViaMistral(ctx, apiKey, text)
	}
	return e, nil
}

func (e *MistralEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return e.embedFn(ctx, text)
}

func (e *MistralEmbedder) Close() {}

// EmbedEmail concatenates email fields and embeds them.
// Returns the embedding vector and the combined text (used for storing in email_history).
func (e *MistralEmbedder) EmbedEmail(ctx context.Context, from, to, subject, body string) ([]float32, string, error) {
	combined := fmt.Sprintf("FROM: %s\nTO: %s\nSUBJECT: %s\nBODY: %s", from, to, subject, body)
	vec, err := e.embedFn(ctx, combined)
	if err != nil {
		return nil, "", err
	}
	return vec, combined, nil
}

func embedViaMistral(ctx context.Context, apiKey, text string) ([]float32, error) {
	const url = "https://api.mistral.ai/v1/embeddings"
	payload, err := json.Marshal(map[string]any{
		"model": "mistral-embed",
		"input": []string{text},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return nil, fmt.Errorf("embedding API: %s", result.Error.Message)
		}
		return nil, fmt.Errorf("embedding API: status %d", resp.StatusCode)
	}
	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return result.Data[0].Embedding, nil
}
