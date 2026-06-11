package main

import (
	"context"
	"log"
	"time"
)

// startRetentionJob runs a daily cleanup that deletes audit_log and quarantine
// rows older than each tenant's configured retention_days (default 90).
func (app *Config) startRetentionJob(ctx context.Context) {
	// Run once at startup, then every 24 hours.
	app.runRetentionOnce(ctx)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			app.runRetentionOnce(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (app *Config) runRetentionOnce(ctx context.Context) {
	deleted, err := app.Store.RunRetention(ctx)
	if err != nil {
		log.Printf("[retention] error: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("[retention] deleted %d expired rows", deleted)
	}
}
