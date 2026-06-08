package main

import "net/http"

func (app *Config) APIKeyMiddleware(next http.Handler) http.Handler {
	return next
}
