package main

import (
	"encoding/json"
	"log"
	"net/http"

	gmailapi "google.golang.org/api/gmail/v1"
)

// GmailArchiveCallback is called by ai-compliance-service after quarantining a Gmail message.
// It removes the email from the user's inbox and sends a quarantine notification.
func (app *Config) GmailArchiveCallback(w http.ResponseWriter, r *http.Request) {
	if app.InternalSecret != "" && r.Header.Get("X-Internal-Secret") != app.InternalSecret {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		UserID         string `json:"user_id"`
		GmailMessageID string `json:"gmail_message_id"`
		Verdict        string `json:"verdict"`
		To             string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.UserID == "" || req.GmailMessageID == "" {
		http.Error(w, "user_id and gmail_message_id required", http.StatusBadRequest)
		return
	}

	go func() {
		ctx := r.Context()
		tok, err := app.Store.GetOAuthToken(ctx, req.UserID, "google")
		if err != nil {
			log.Printf("[archive-callback] no oauth token for user %s: %v", req.UserID, err)
			return
		}

		svc, _, err := app.buildGmailClient(ctx, tok)
		if err != nil {
			log.Printf("[archive-callback] build gmail client: %v", err)
			return
		}

		_, err = svc.Users.Messages.Modify("me", req.GmailMessageID, &gmailapi.ModifyMessageRequest{
			RemoveLabelIds: []string{"INBOX"},
		}).Do()
		if err != nil {
			log.Printf("[archive-callback] remove from inbox: %v", err)
		}

		if req.To != "" {
			go app.sendQuarantineNotification(req.To, "", "")
		}
	}()

	w.WriteHeader(http.StatusOK)
}
