package handler

import (
	"net/http"
	"net/url"

	"github.com/coder/websocket"

	"github.com/miraccan/discord-backend/internal/auth"
	"github.com/miraccan/discord-backend/internal/signaling"
)

// WS upgrades authenticated requests to a signaling WebSocket and hands them to
// the hub.
type WS struct {
	hub            *signaling.Hub
	issuer         *auth.Issuer
	originPatterns []string
	skipOrigin     bool
}

// NewWS builds the WebSocket handler. allowedOrigins are full origin URLs (e.g.
// "http://localhost:5173"); a single "*" disables origin verification.
func NewWS(hub *signaling.Hub, issuer *auth.Issuer, allowedOrigins []string, allowAll bool) *WS {
	return &WS{
		hub:            hub,
		issuer:         issuer,
		originPatterns: originHosts(allowedOrigins),
		skipOrigin:     allowAll,
	}
}

// ServeHTTP handles GET /ws?token=<jwt>. The browser WebSocket API cannot set
// an Authorization header, so the token travels as a query parameter.
func (h *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	username, err := h.issuer.Verify(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:     h.originPatterns,
		InsecureSkipVerify: h.skipOrigin,
	})
	if err != nil {
		return // Accept already wrote the response.
	}
	defer func() {
		_ = conn.CloseNow() //nolint:errcheck // best-effort close on handler exit
	}()

	h.hub.Serve(r.Context(), conn, username)
}

// originHosts converts origin URLs into the host[:port] patterns that the
// WebSocket library matches against the request's Origin header.
func originHosts(origins []string) []string {
	hosts := make([]string, 0, len(origins))
	for _, o := range origins {
		if o == "*" {
			continue
		}
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			hosts = append(hosts, u.Host)
		} else {
			hosts = append(hosts, o)
		}
	}
	return hosts
}
