package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const (
	contextKeyTenantID contextKey = "tenant_id"
	contextKeyUserID   contextKey = "user_id"
	contextKeyRole     contextKey = "role"
	contextKeyEmail    contextKey = "email"
)

// APIKeyMiddleware authenticates requests via Bearer API key and sets tenant_id in context.
func (app *Config) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			app.errorJSON(w, fmt.Errorf("missing or invalid Authorization header"), http.StatusUnauthorized)
			return
		}
		rawKey := strings.TrimPrefix(authHeader, "Bearer ")

		tenantID, err := app.Store.ValidateAPIKey(r.Context(), rawKey)
		if err != nil {
			app.errorJSON(w, fmt.Errorf("invalid API key"), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyTenantID, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWTMiddleware authenticates requests via Bearer JWT and sets user_id, tenant_id, role in context.
func (app *Config) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			app.errorJSON(w, fmt.Errorf("missing or invalid Authorization header"), http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := app.parseAccessToken(tokenStr)
		if err != nil {
			app.errorJSON(w, fmt.Errorf("invalid or expired token"), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyTenantID, claims.TenantID)
		ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
		ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FlexAuthMiddleware accepts either a JWT access token or an API key.
// JWT: sets user_id, tenant_id, role. API key: sets tenant_id only.
func (app *Config) FlexAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			app.errorJSON(w, fmt.Errorf("missing or invalid Authorization header"), http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Try JWT first
		if claims, err := app.parseAccessToken(token); err == nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyTenantID, claims.TenantID)
			ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
			ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Fall back to API key
		tenantID, err := app.Store.ValidateAPIKey(r.Context(), token)
		if err != nil {
			app.errorJSON(w, fmt.Errorf("invalid credentials"), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyTenantID, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that allows only the specified roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(contextKeyRole).(string)
			if !allowed[role] {
				http.Error(w, `{"error":true,"message":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
