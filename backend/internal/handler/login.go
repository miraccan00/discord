package handler

import (
	"encoding/json"
	"net/http"

	"github.com/miraccan/discord-backend/internal/auth"
)

// Login authenticates a user against the dummy credential set and returns a JWT.
type Login struct {
	issuer *auth.Issuer
}

// NewLogin builds the login handler.
func NewLogin(issuer *auth.Issuer) *Login {
	return &Login{issuer: issuer}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// ServeHTTP handles POST /api/login.
func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if !auth.CheckCredentials(req.Username, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	token, err := l.issuer.Issue(req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not issue token"})
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token, Username: req.Username})
}
