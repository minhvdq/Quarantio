package embeddings

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiEmbedder struct {
	client *genai.Client
	model  *genai.EmbeddingModel
}

func NewGeminiEmbedder(ctx context.Context, apiKey string) (*GeminiEmbedder, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiEmbedder{
		client: client,
		model:  client.EmbeddingModel("text-embedding-004"),
	}, nil
}

func (g *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	res, err := g.model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}
	return res.Embedding.Values, nil
}

func (g *GeminiEmbedder) Close() {
	g.client.Close()
}
