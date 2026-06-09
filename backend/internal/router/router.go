// Package router wires HTTP routes and middleware for the signaling server.
package router

import (
	"net/http"
	"slices"

	"github.com/miraccan/discord-backend/internal/auth"
	"github.com/miraccan/discord-backend/internal/config"
	"github.com/miraccan/discord-backend/internal/handler"
	"github.com/miraccan/discord-backend/internal/signaling"
)

// New builds the HTTP handler for the whole API surface.
func New(cfg config.Config, hub *signaling.Hub, issuer *auth.Issuer) http.Handler {
	mux := http.NewServeMux()

	login := handler.NewLogin(issuer)
	ws := handler.NewWS(hub, issuer, cfg.AllowedOrigins, cfg.AllowAllOrigins())

	mux.Handle("/api/login", cors(cfg, login))
	mux.Handle("/ws", ws)
	mux.HandleFunc("/healthz", handler.Healthz)
	mux.HandleFunc("/ready", handler.Ready)

	return mux
}

// cors adds permissive-but-scoped CORS headers for the JSON API and answers
// preflight requests.
func cors(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (cfg.AllowAllOrigins() || slices.Contains(cfg.AllowedOrigins, origin)) {
			allow := origin
			if cfg.AllowAllOrigins() {
				allow = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", allow)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
