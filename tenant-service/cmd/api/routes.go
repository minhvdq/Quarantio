package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))
	mux.Use(middleware.Logger)

	// Public auth endpoints
	mux.Post("/auth/register", app.Register)
	mux.Post("/auth/login", app.Login)
	mux.Post("/auth/refresh", app.Refresh)
	mux.Post("/auth/logout", app.Logout)

	// Legacy: org registration via API (creates tenant + API key)
	mux.Post("/v1/organizations", app.RegisterOrganization)

	// Flex-auth routes: accept JWT or API key (backward-compatible)
	mux.Group(func(r chi.Router) {
		r.Use(app.FlexAuthMiddleware)
		r.Post("/v1/check", app.CheckEmail)
		r.Post("/v1/policies", app.UploadPolicy)
		r.Get("/v1/audit", app.GetAuditLog)
		r.Get("/v1/quarantine", app.GetQuarantine)
		r.Post("/v1/quarantine/{id}/review", app.ReviewQuarantine)
		r.Get("/v1/policies", app.ListPolicies)
		r.Delete("/v1/policies", app.DeletePolicy)
		r.Get("/v1/settings", app.GetSettings)
		r.Post("/v1/settings", app.UpdateSettings)
		r.Get("/v1/export", app.ExportData)
		r.Delete("/v1/data", app.DeleteData)
	})

	// JWT authenticated routes (dashboard users)
	mux.Group(func(r chi.Router) {
		r.Use(app.JWTMiddleware)

		r.Get("/v1/me", app.Me)

		// Member management — owner only
		r.With(RequireRole("owner")).Get("/v1/members", app.ListMembers)
		r.With(RequireRole("owner")).Post("/v1/members", app.InviteMember)
		r.With(RequireRole("owner")).Patch("/v1/members/{id}/role", app.UpdateMemberRole)
		r.With(RequireRole("owner")).Delete("/v1/members/{id}", app.RemoveMember)

		// Any user can submit a HIGH release request; only owner can action it
		r.Post("/v1/quarantine/{id}/release-request", app.SubmitReleaseRequest)
		r.With(RequireRole("owner")).Get("/v1/release-requests", app.ListReleaseRequests)
		r.With(RequireRole("owner")).Post("/v1/release-requests/{id}/action", app.ActionReleaseRequest)
	})

	return mux
}
