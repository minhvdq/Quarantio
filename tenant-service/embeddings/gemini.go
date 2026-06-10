package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GeminiEmbedder struct {
	apiKey string
}

// NewGeminiEmbedder creates an embedder using the Mistral embeddings API.
// The type name is kept for compatibility with existing handler references.
func NewGeminiEmbedder(_ context.Context, apiKey string) (*GeminiEmbedder, error) {
	return &GeminiEmbedder{apiKey: apiKey}, nil
}

func (g *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
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
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

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

func (g *GeminiEmbedder) Close() {}
