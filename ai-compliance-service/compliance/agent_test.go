package compliance

import (
	"context"
	"encoding/json"
	"testing"
)

type mockRAGStore struct {
	policyChunks  []RAGChunk
	historyChunks []RAGChunk
}

func (m *mockRAGStore) QueryPolicyChunks(_ context.Context, _ string, _ []float32, _ int) ([]RAGChunk, error) {
	return m.policyChunks, nil
}

func (m *mockRAGStore) QueryHistoryChunks(_ context.Context, _ string, _ []float32, _ int) ([]RAGChunk, error) {
	return m.historyChunks, nil
}

func newTestAgent(store RAGStore) *GeminiAgent {
	e := &GeminiEmbedder{}
	e.embedFn = func(_ context.Context, _ string) ([]float32, error) {
		return []float32{0.1, 0.2, 0.3}, nil
	}
	return &GeminiAgent{embedder: e, store: store}
}

func TestToolCheckPolicyViolation_WithChunks(t *testing.T) {
	store := &mockRAGStore{
		policyChunks: []RAGChunk{
			{Source: "hr-policy.pdf", Content: "No PII in external emails"},
		},
	}
	agent := newTestAgent(store)
	result := agent.toolCheckPolicyViolation(context.Background(), "send SSN to client", "tenant-1")

	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if got["policy_match"] != true {
		t.Errorf("expected policy_match=true, got %v", got["policy_match"])
	}
}

func TestToolCheckPolicyViolation_NoChunks(t *testing.T) {
	store := &mockRAGStore{policyChunks: []RAGChunk{}}
	agent := newTestAgent(store)
	result := agent.toolCheckPolicyViolation(context.Background(), "hello world", "tenant-1")

	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	if got["policy_match"] != false {
		t.Errorf("expected policy_match=false, got %v", got["policy_match"])
	}
}

func TestToolCheckPolicyViolation_NoTenant(t *testing.T) {
	agent := newTestAgent(nil)
	result := agent.toolCheckPolicyViolation(context.Background(), "hello", "")

	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["policy_match"] != false {
		t.Errorf("expected policy_match=false, got %v", got["policy_match"])
	}
	if got["reason"] == nil {
		t.Errorf("expected reason field in no-tenant response")
	}
}

func TestToolRetrievePrecedent_WithChunks(t *testing.T) {
	store := &mockRAGStore{
		historyChunks: []RAGChunk{
			{Source: "HIGH", Content: "Email attempted to exfiltrate customer data"},
		},
	}
	agent := newTestAgent(store)
	result := agent.toolRetrievePrecedent(context.Background(), "send all customer records", "tenant-1")

	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v — got: %s", err, result)
	}
	precedents, ok := got["precedents"].([]any)
	if !ok || len(precedents) == 0 {
		t.Errorf("expected non-empty precedents, got %v", got["precedents"])
	}
}

func TestToolRetrievePrecedent_NoTenant(t *testing.T) {
	agent := newTestAgent(nil)
	result := agent.toolRetrievePrecedent(context.Background(), "hello", "")

	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	precedents, ok := got["precedents"].([]any)
	if !ok || len(precedents) != 0 {
		t.Errorf("expected empty precedents array, got %v", got["precedents"])
	}
}
