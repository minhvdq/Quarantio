package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *Config) processEmail(
	ctx context.Context,
	email EmailMessage,
	agent AgentRunner,
	embedder Embedder,
	pub Publisher,
) error {
	combined := fmt.Sprintf("FROM: %s\nTO: %s\nSUBJECT: %s\nBODY: %s",
		email.From, email.To, email.Subject, email.Message)

	log.Println("[pipeline] embedding email...")
	t0 := time.Now()
	vec, err := embedder.Embed(ctx, combined)
	if err != nil {
		return fmt.Errorf("embed email: %w", err)
	}
	log.Printf("[pipeline] embedded in %s", time.Since(t0).Round(time.Millisecond))

	var policyChunks, historyChunks []RAGChunk
	if email.TenantID != "" {
		log.Println("[pipeline] querying RAG chunks...")
		t1 := time.Now()
		policyChunks, err = app.Store.QueryPolicyChunks(ctx, email.TenantID, vec, 5)
		if err != nil {
			return fmt.Errorf("query policy: %w", err)
		}
		historyChunks, err = app.Store.QueryHistoryChunks(ctx, email.TenantID, vec, 3)
		if err != nil {
			return fmt.Errorf("query history: %w", err)
		}
		log.Printf("[pipeline] RAG: %d policy chunks, %d history chunks (took %s)", len(policyChunks), len(historyChunks), time.Since(t1).Round(time.Millisecond))
	}

	log.Println("[pipeline] running agent loop...")
	t2 := time.Now()
	decision, err := agent.RunLoop(ctx, email, policyChunks, historyChunks)
	if err != nil {
		return fmt.Errorf("agent loop: %w", err)
	}
	log.Printf("[pipeline] agent verdict=%s violations=%v (took %s)", decision.Verdict, decision.Violations, time.Since(t2).Round(time.Millisecond))

	settings := &TenantSettings{AutoDeliverLow: true, RetentionDays: 90}
	if email.TenantID != "" {
		if s, err := app.Store.GetTenantSettings(ctx, email.TenantID); err == nil {
			settings = s
		}
	}

	action := verdictAction(decision.Verdict)

	switch decision.Verdict {
	case VerdictClean:
		if app.MailServiceURL != "" {
			if err := app.sendToMailService(ctx, email, email.Message); err != nil {
				return fmt.Errorf("send to mail-service: %w", err)
			}
		}

	case VerdictLow:
		if settings.AutoDeliverLow {
			body := email.Message
			if decision.RemediatedBody != "" {
				body = decision.RemediatedBody
			}
			if app.MailServiceURL != "" {
				if err := app.sendToMailService(ctx, email, body); err != nil {
					return fmt.Errorf("send to mail-service: %w", err)
				}
			}
		} else {
			action = "quarantined"
			payload, _ := json.Marshal(email)
			if err := pub.Publish(ctx, payload, "email.quarantine"); err != nil {
				return fmt.Errorf("publish quarantine (low): %w", err)
			}
		}

	case VerdictMedium:
		msg := email
		msg.Violations = decision.Violations
		msg.Reasoning = decision.Reasoning
		payload, _ := json.Marshal(msg)
		if err := pub.Publish(ctx, payload, "email.quarantine"); err != nil {
			return fmt.Errorf("publish quarantine: %w", err)
		}

	case VerdictHigh:
		msg := email
		msg.Violations = decision.Violations
		msg.Reasoning = decision.Reasoning
		payload, _ := json.Marshal(msg)
		if err := pub.Publish(ctx, payload, "email.blocked"); err != nil {
			return fmt.Errorf("publish blocked: %w", err)
		}
	}

	entry := AuditEntry{
		TenantID:   email.TenantID,
		EmailFrom:  email.From,
		EmailTo:    []string{email.To},
		Subject:    email.Subject,
		Verdict:    decision.Verdict,
		Violations: decision.Violations,
		Reasoning:  decision.Reasoning,
		Action:     action,
	}
	if err := app.Store.InsertAuditLog(ctx, entry); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	if decision.Verdict == VerdictClean || decision.Verdict == VerdictLow {
		if err := app.Store.InsertEmailHistory(ctx, email.TenantID, combined, vec, decision.Verdict, decision.Violations); err != nil {
			log.Printf("InsertEmailHistory (best-effort): %v", err)
		}
	}

	return nil
}

func (app *Config) sendToMailService(ctx context.Context, email EmailMessage, body string) error {
	payload, err := json.Marshal(map[string]string{
		"to":      email.To,
		"subject": email.Subject,
		"message": body,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", app.MailServiceURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mail-service returned %d", resp.StatusCode)
	}
	return nil
}

func verdictAction(v Verdict) string {
	switch v {
	case VerdictClean:
		return "delivered"
	case VerdictLow:
		return "remediated_and_delivered"
	case VerdictMedium:
		return "quarantined"
	case VerdictHigh:
		return "blocked"
	default:
		return "unknown"
	}
}

// checkEmail runs embed → RAG → agent → audit without routing to any queue.
// Used by the sync check HTTP endpoint.
func (app *Config) checkEmail(ctx context.Context, email EmailMessage, agent AgentRunner, embedder Embedder) (*Decision, error) {
	combined := fmt.Sprintf("FROM: %s\nTO: %s\nSUBJECT: %s\nBODY: %s",
		email.From, email.To, email.Subject, email.Message)

	vec, err := embedder.Embed(ctx, combined)
	if err != nil {
		return nil, fmt.Errorf("embed email: %w", err)
	}

	var policyChunks, historyChunks []RAGChunk
	if email.TenantID != "" {
		policyChunks, err = app.Store.QueryPolicyChunks(ctx, email.TenantID, vec, 5)
		if err != nil {
			return nil, fmt.Errorf("query policy: %w", err)
		}
		historyChunks, err = app.Store.QueryHistoryChunks(ctx, email.TenantID, vec, 3)
		if err != nil {
			return nil, fmt.Errorf("query history: %w", err)
		}
	}

	decision, err := agent.RunLoop(ctx, email, policyChunks, historyChunks)
	if err != nil {
		return nil, fmt.Errorf("agent loop: %w", err)
	}

	entry := AuditEntry{
		TenantID:   email.TenantID,
		EmailFrom:  email.From,
		EmailTo:    []string{email.To},
		Subject:    email.Subject,
		Verdict:    decision.Verdict,
		Violations: decision.Violations,
		Reasoning:  decision.Reasoning,
		Action:     "checked",
	}
	if err := app.Store.InsertAuditLog(ctx, entry); err != nil {
		log.Printf("[check] audit log (best-effort): %v", err)
	}

	return decision, nil
}

// handleSyncCheck handles POST /internal/check — runs the compliance agent and returns the verdict.
func (app *Config) handleSyncCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var email EmailMessage
	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	decision, err := app.checkEmail(r.Context(), email, app.Agent, app.Embedder)
	if err != nil {
		log.Printf("[check] error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"verdict":         string(decision.Verdict),
		"violations":      decision.Violations,
		"reasoning":       decision.Reasoning,
		"remediated_body": decision.RemediatedBody,
	})
}

func (app *Config) runQuarantineWorker(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var email EmailMessage
		if err := json.Unmarshal(d.Body, &email); err != nil {
			log.Printf("[quarantine] unmarshal error: %v — nacking", err)
			_ = d.Nack(false, false)
			continue
		}
		vJSON := violationsJSON(email.Violations)
		if err := app.Store.InsertQuarantine(
			context.Background(),
			email.TenantID, email.From, email.To,
			email.Subject, email.Message,
			vJSON, email.Reasoning, "medium",
		); err != nil {
			log.Printf("[quarantine] insert error: %v — nacking", err)
			_ = d.Nack(false, true)
			continue
		}
		log.Printf("[quarantine] stored for review: from=%s subject=%q", email.From, email.Subject)
		if email.GmailMessageID != "" && app.TenantSvcURL != "" {
			go app.callGmailArchive(email.UserID, email.GmailMessageID, "MEDIUM", email.To)
		}
		_ = d.Ack(false)
	}
}

func (app *Config) runBlockedWorker(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var email EmailMessage
		if err := json.Unmarshal(d.Body, &email); err != nil {
			log.Printf("[blocked] unmarshal error: %v — nacking", err)
			_ = d.Nack(false, false)
			continue
		}
		vJSON := violationsJSON(email.Violations)
		if err := app.Store.InsertQuarantine(
			context.Background(),
			email.TenantID, email.From, email.To,
			email.Subject, email.Message,
			vJSON, email.Reasoning, "high",
		); err != nil {
			log.Printf("[blocked] insert quarantine error: %v — nacking", err)
			_ = d.Nack(false, true)
			continue
		}
		log.Printf("[blocked] quarantined HIGH email: from=%s subject=%q", email.From, email.Subject)
		if email.GmailMessageID != "" && app.TenantSvcURL != "" {
			go app.callGmailArchive(email.UserID, email.GmailMessageID, "HIGH", email.To)
		}
		_ = d.Ack(false)
	}
}

func violationsJSON(v []string) string {
	if len(v) == 0 {
		return `[]`
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (app *Config) callGmailArchive(userID, gmailMessageID, verdict, to string) {
	payload, _ := json.Marshal(map[string]string{
		"user_id":          userID,
		"gmail_message_id": gmailMessageID,
		"verdict":          verdict,
		"to":               to,
	})
	resp, err := http.Post(
		app.TenantSvcURL+"/internal/gmail/archive",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		log.Printf("[archive-callback] failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[archive-callback] tenant-service returned %d", resp.StatusCode)
	}
}
