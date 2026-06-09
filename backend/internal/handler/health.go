// Package handler contains the HTTP handlers for the signaling server.
package handler

import (
	"encoding/json"
	"net/http"
)

// writeJSON serializes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// A failed write means the client disconnected mid-response; nothing here
	// can recover from that, so the error is intentionally dropped.
	_, _ = w.Write(data) // client disconnected mid-response; nothing to recover
}

// Healthz is the Kubernetes liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

// Ready is the Kubernetes readiness probe.
func Ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
