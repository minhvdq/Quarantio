package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (app *Config) ListMembers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(contextKeyTenantID).(string)

	members, err := app.Store.ListOrgMembers(r.Context(), tenantID)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("list members: %w", err), http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusOK, members)
}

type inviteMemberRequest struct {
	Email string `json:"email"`
}

// InviteMember adds a user to the org. All invited members join as "user" role.
// Owners are set at registration and cannot be assigned via invite.
func (app *Config) InviteMember(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(contextKeyTenantID).(string)
	inviterID := r.Context().Value(contextKeyUserID).(string)

	var req inviteMemberRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorJSON(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		app.errorJSON(w, errors.New("email is required"))
		return
	}

	ctx := r.Context()

	user, err := app.Store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			app.errorJSON(w, fmt.Errorf("lookup user: %w", err), http.StatusInternalServerError)
			return
		}
		user, err = app.Store.CreateUser(ctx, req.Email, tempPassword(), "", "")
		if err != nil {
			app.errorJSON(w, fmt.Errorf("create invited user: %w", err), http.StatusInternalServerError)
			return
		}
	}

	if err := app.Store.CreateOrgMember(ctx, user.ID, tenantID, "user", &inviterID); err != nil {
		if strings.Contains(err.Error(), "unique") {
			app.errorJSON(w, errors.New("user is already a member"), http.StatusConflict)
		} else {
			app.errorJSON(w, fmt.Errorf("add member: %w", err), http.StatusInternalServerError)
		}
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]string{"message": "member added", "user_id": user.ID})
}

func (app *Config) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	app.errorJSON(w, errors.New("roles are fixed: owner or user"), http.StatusBadRequest)
}

func (app *Config) RemoveMember(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(contextKeyTenantID).(string)
	memberID := chi.URLParam(r, "id")

	if err := app.Store.RemoveOrgMember(r.Context(), memberID, tenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.errorJSON(w, errors.New("member not found or is the owner"), http.StatusNotFound)
		} else {
			app.errorJSON(w, fmt.Errorf("remove member: %w", err), http.StatusInternalServerError)
		}
		return
	}
	app.writeJSON(w, http.StatusOK, jsonResponse{Message: "member removed"})
}

func tempPassword() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
