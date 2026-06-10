package compliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	mistralChatURL = "https://api.mistral.ai/v1/chat/completions"
	mistralModel   = "mistral-small-latest"
)

// verdictJSON is what Mistral must return as its final text response.
type verdictJSON struct {
	Verdict        string   `json:"verdict"`
	Violations     []string `json:"violations"`
	Reasoning      string   `json:"reasoning"`
	RemediatedBody string   `json:"remediated_body"`
}

type policyToolResult struct {
	PolicyMatch bool        `json:"policy_match"`
	Chunks      []chunkItem `json:"chunks,omitempty"`
	Reason      string      `json:"reason,omitempty"`
}

type chunkItem struct {
	Source  string `json:"source"`
	Content string `json:"content"`
}

type precedentToolResult struct {
	Precedents []precedentItem `json:"precedents"`
	Reason     string          `json:"reason,omitempty"`
}

type precedentItem struct {
	Verdict string `json:"verdict"`
	Summary string `json:"summary"`
}

// These types mirror the main package to avoid import cycle.
// The agentAdapter in main.go bridges them.
type EmailMessage struct {
	From, To, Subject, Message, TenantID string
}

type RAGChunk struct {
	Content, Source string
}

type Decision struct {
	Verdict        string
	Violations     []string
	Reasoning      string
	RemediatedBody string
}

// RAGStore provides pgvector similarity queries for in-loop tool calls.
type RAGStore interface {
	QueryPolicyChunks(ctx context.Context, tenantID string, embedding []float32, limit int) ([]RAGChunk, error)
	QueryHistoryChunks(ctx context.Context, tenantID string, embedding []float32, limit int) ([]RAGChunk, error)
}

// --- Mistral HTTP types ---

type mistralMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content,omitempty"`
	ToolCalls  []mistralToolCall `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	Name       string            `json:"name,omitempty"`
}

type mistralToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function mistralFunctionCall `json:"function"`
}

type mistralFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type mistralTool struct {
	Type     string          `json:"type"`
	Function mistralFunction `json:"function"`
}

type mistralFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type mistralRequest struct {
	Model    string           `json:"model"`
	Messages []mistralMessage `json:"messages"`
	Tools    []mistralTool    `json:"tools,omitempty"`
}

type mistralResponse struct {
	Choices []struct {
		Message struct {
			Role      string           `json:"role"`
			Content   string           `json:"content"`
			ToolCalls []mistralToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// MistralAgent runs a multi-turn Mistral function-calling compliance loop.
type MistralAgent struct {
	apiKey   string
	embedder *MistralEmbedder
	store    RAGStore
}

func NewMistralAgent(_ context.Context, apiKey string, store RAGStore) (*MistralAgent, error) {
	embedder, err := NewMistralEmbedder(context.Background(), apiKey)
	if err != nil {
		return nil, err
	}
	return &MistralAgent{apiKey: apiKey, embedder: embedder, store: store}, nil
}

func (a *MistralAgent) Close() {}

// RunLoop runs the Mistral agent loop and returns a compliance Decision.
// policyChunks and historyChunks are pre-fetched RAG context injected into the system prompt.
func (a *MistralAgent) RunLoop(ctx context.Context, email EmailMessage, policyChunks, historyChunks []RAGChunk) (*Decision, error) {
	tools := buildMistralTools()
	messages := []mistralMessage{
		{Role: "system", Content: buildSystemPrompt(policyChunks, historyChunks)},
		{Role: "user", Content: fmt.Sprintf(
			"Analyze this email for compliance violations:\n\nFROM: %s\nTO: %s\nSUBJECT: %s\nBODY:\n%s",
			email.From, email.To, email.Subject, email.Message,
		)},
	}

	for i := 0; i < 10; i++ {
		resp, err := sendMistralWithRetry(ctx, a.apiKey, mistralRequest{
			Model:    mistralModel,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, fmt.Errorf("mistral chat: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("mistral returned no choices")
		}
		choice := resp.Choices[0]

		if len(choice.Message.ToolCalls) > 0 {
			messages = append(messages, mistralMessage{
				Role:      "assistant",
				ToolCalls: choice.Message.ToolCalls,
			})
			for _, tc := range choice.Message.ToolCalls {
				var args map[string]any
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				result := a.executeTool(ctx, tc.Function.Name, args, email.TenantID)
				messages = append(messages, mistralMessage{
					Role:       "tool",
					Content:    result,
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
				})
			}
			continue
		}

		return parseVerdict(choice.Message.Content)
	}

	return nil, fmt.Errorf("agent loop exceeded max iterations")
}

func sendMistralWithRetry(ctx context.Context, apiKey string, req mistralRequest) (*mistralResponse, error) {
	backoff := 10 * time.Second
	for attempt := 0; attempt < 4; attempt++ {
		resp, err := sendMistralOnce(ctx, apiKey, req)
		if err == nil {
			return resp, nil
		}
		if !strings.Contains(err.Error(), "429") {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	return sendMistralOnce(ctx, apiKey, req)
}

func sendMistralOnce(ctx context.Context, apiKey string, req mistralRequest) (*mistralResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", mistralChatURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("mistral 429: rate limited")
	}

	var result mistralResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode mistral response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("mistral error: %s", result.Error.Message)
	}
	return &result, nil
}

func buildSystemPrompt(policy, history []RAGChunk) string {
	var sb strings.Builder
	sb.WriteString("You are a compliance officer AI reviewing outbound emails.\n\n")

	if len(policy) > 0 {
		sb.WriteString("TENANT POLICY CONTEXT:\n")
		for _, c := range policy {
			sb.WriteString(fmt.Sprintf("[%s]: %s\n", c.Source, c.Content))
		}
		sb.WriteString("\n")
	}

	if len(history) > 0 {
		sb.WriteString("HISTORICAL PRECEDENTS (verdict — email summary):\n")
		for _, c := range history {
			sb.WriteString(fmt.Sprintf("[%s]: %s\n", c.Source, c.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Use available tools to investigate. When finished, respond with ONLY this JSON (no markdown):
{"verdict":"CLEAN|LOW|MEDIUM|HIGH","violations":["..."],"reasoning":"...","remediated_body":"..."}

Verdicts:
- CLEAN: no violations found
- LOW: minor issue, auto-remediable (populate remediated_body with cleaned version)
- MEDIUM: ambiguous risk requiring human review
- HIGH: clear threat (phishing, exfiltration, severe PII leak)`)

	return sb.String()
}

func schemaObj(props map[string]string, required []string) json.RawMessage {
	properties := make(map[string]map[string]string, len(props))
	for k, desc := range props {
		properties[k] = map[string]string{"type": "string", "description": desc}
	}
	b, _ := json.Marshal(map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	})
	return b
}

func buildMistralTools() []mistralTool {
	tool := func(name, desc string, params json.RawMessage) mistralTool {
		return mistralTool{Type: "function", Function: mistralFunction{Name: name, Description: desc, Parameters: params}}
	}
	return []mistralTool{
		tool("scan_pii", "Scan text for PII: SSNs, credit card numbers, phone numbers",
			schemaObj(map[string]string{"content": "Text to scan for PII"}, []string{"content"})),
		tool("check_phishing", "Detect phishing signals: urgency language, credential requests, spoofed sender, lookalike domains",
			schemaObj(map[string]string{"content": "Email body text", "sender": "Sender email address"}, []string{"content", "sender"})),
		tool("check_policy_violation", "RAG search against the tenant's uploaded compliance policies",
			schemaObj(map[string]string{"content": "Text to check against policy"}, []string{"content"})),
		tool("check_exfiltration", "Flag data exfiltration: bulk recipients, encoded content, confidential leaks",
			schemaObj(map[string]string{"recipients": "Comma-separated recipient addresses", "content": "Email body text"}, []string{"recipients", "content"})),
		tool("retrieve_precedent", "RAG search against historical approved/flagged emails for similar past verdents",
			schemaObj(map[string]string{"content": "Email content to find precedents for"}, []string{"content"})),
		tool("remediate_content", "Rewrite or redact email body to remove LOW-severity violations while preserving intent",
			schemaObj(map[string]string{"content": "Original email body", "violations": "Comma-separated violations to remediate"}, []string{"content", "violations"})),
	}
}

func (a *MistralAgent) executeTool(ctx context.Context, name string, args map[string]any, tenantID string) string {
	str := func(key string) string {
		v, _ := args[key].(string)
		return v
	}
	switch name {
	case "scan_pii":
		return toolScanPII(str("content"))
	case "check_phishing":
		return toolCheckPhishing(str("content"), str("sender"))
	case "check_policy_violation":
		return a.toolCheckPolicyViolation(ctx, str("content"), tenantID)
	case "check_exfiltration":
		return toolCheckExfiltration(str("recipients"), str("content"))
	case "retrieve_precedent":
		return a.toolRetrievePrecedent(ctx, str("content"), tenantID)
	case "remediate_content":
		return a.toolRemediateContent(ctx, str("content"), str("violations"))
	default:
		return `{"error":"unknown tool"}`
	}
}

var (
	reSSN   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	reCC    = regexp.MustCompile(`\b(?:\d{4}[- ]?){3}\d{4}\b`)
	rePhone = regexp.MustCompile(`\b\d{3}[.\-]\d{3}[.\-]\d{4}\b`)
)

func toolScanPII(content string) string {
	var found []string
	if reSSN.MatchString(content) {
		found = append(found, "SSN pattern detected")
	}
	if reCC.MatchString(content) {
		found = append(found, "credit card pattern detected")
	}
	if rePhone.MatchString(content) {
		found = append(found, "phone number detected")
	}
	if len(found) == 0 {
		return `{"pii_found":false,"details":[]}`
	}
	details, _ := json.Marshal(found)
	return fmt.Sprintf(`{"pii_found":true,"details":%s}`, details)
}

func toolCheckPhishing(content, sender string) string {
	urgencyWords := []string{"urgent", "immediate", "verify now", "account suspended", "click here", "login required"}
	var signals []string
	lower := strings.ToLower(content + " " + sender)
	for _, w := range urgencyWords {
		if strings.Contains(lower, w) {
			signals = append(signals, w)
		}
	}
	if strings.Contains(lower, "paypa1") || strings.Contains(lower, "arnazon") || strings.Contains(lower, "micros0ft") {
		signals = append(signals, "lookalike domain detected")
	}
	if len(signals) == 0 {
		return `{"phishing_risk":"low","signals":[]}`
	}
	s, _ := json.Marshal(signals)
	return fmt.Sprintf(`{"phishing_risk":"high","signals":%s}`, s)
}

func (a *MistralAgent) toolCheckPolicyViolation(ctx context.Context, content, tenantID string) string {
	if tenantID == "" || a.store == nil || a.embedder == nil {
		out, _ := json.Marshal(policyToolResult{PolicyMatch: false, Reason: "no tenant context"})
		return string(out)
	}
	vec, err := a.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	chunks, err := a.store.QueryPolicyChunks(ctx, tenantID, vec, 3)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	items := make([]chunkItem, len(chunks))
	for i, c := range chunks {
		items[i] = chunkItem{Source: c.Source, Content: c.Content}
	}
	out, _ := json.Marshal(policyToolResult{PolicyMatch: len(chunks) > 0, Chunks: items})
	return string(out)
}

func toolCheckExfiltration(recipients, content string) string {
	recipientList := strings.Split(recipients, ",")
	var signals []string
	if len(recipientList) > 20 {
		signals = append(signals, fmt.Sprintf("bulk send: %d recipients", len(recipientList)))
	}
	lower := strings.ToLower(content)
	if strings.Contains(lower, "confidential") && len(recipientList) > 5 {
		signals = append(signals, "confidential content to large distribution")
	}
	if len(signals) == 0 {
		return `{"exfiltration_risk":"low","signals":[]}`
	}
	s, _ := json.Marshal(signals)
	return fmt.Sprintf(`{"exfiltration_risk":"high","signals":%s}`, s)
}

func (a *MistralAgent) toolRetrievePrecedent(ctx context.Context, content, tenantID string) string {
	if tenantID == "" || a.store == nil || a.embedder == nil {
		out, _ := json.Marshal(precedentToolResult{Precedents: []precedentItem{}, Reason: "no tenant context"})
		return string(out)
	}
	vec, err := a.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	chunks, err := a.store.QueryHistoryChunks(ctx, tenantID, vec, 3)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	items := make([]precedentItem, len(chunks))
	for i, c := range chunks {
		items[i] = precedentItem{Verdict: c.Source, Summary: c.Content}
	}
	out, _ := json.Marshal(precedentToolResult{Precedents: items})
	return string(out)
}

func (a *MistralAgent) toolRemediateContent(ctx context.Context, content, violations string) string {
	prompt := fmt.Sprintf(
		"Rewrite this email to remove the following violations while preserving the original intent. Return ONLY the rewritten email body.\n\nVIOLATIONS: %s\n\nORIGINAL:\n%s",
		violations, content,
	)
	resp, err := sendMistralOnce(ctx, a.apiKey, mistralRequest{
		Model:    mistralModel,
		Messages: []mistralMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	if len(resp.Choices) == 0 {
		return `{"error":"no response from remediation model"}`
	}
	return fmt.Sprintf(`{"remediated_body":%q}`, resp.Choices[0].Message.Content)
}

func parseVerdict(text string) (*Decision, error) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx > 0 {
		text = text[idx:]
	}
	if idx := strings.LastIndex(text, "}"); idx >= 0 && idx < len(text)-1 {
		text = text[:idx+1]
	}

	var v verdictJSON
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return nil, fmt.Errorf("parseVerdict: %w — raw: %q", err, text)
	}

	return &Decision{
		Verdict:        v.Verdict,
		Violations:     v.Violations,
		Reasoning:      v.Reasoning,
		RemediatedBody: v.RemediatedBody,
	}, nil
}
