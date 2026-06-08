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
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))
	mux.Use(middleware.Logger)

	mux.Post("/v1/organizations", app.RegisterOrganization)

	mux.Group(func(r chi.Router) {
		r.Use(app.APIKeyMiddleware)
		r.Post("/v1/policies", app.UploadPolicy)
	})

	return mux
}
