package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miraccan/discord-backend/internal/auth"
)

func TestLoginSuccess(t *testing.T) {
	// Arrange
	h := NewLogin(auth.NewIssuer("secret", 60))
	body := strings.NewReader(`{"username":"alice","password":"alice123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("response missing token: %s", rec.Body.String())
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	// Arrange
	h := NewLogin(auth.NewIssuer("secret", 60))
	body := strings.NewReader(`{"username":"alice","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLoginRejectsGet(t *testing.T) {
	// Arrange
	h := NewLogin(auth.NewIssuer("secret", 60))
	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
