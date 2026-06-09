// Command server runs the WebRTC signaling backend: a JSON-over-WebSocket relay
// that brokers room membership and SDP/ICE exchange for browser peers. No media
// passes through this process — audio flows peer-to-peer.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miraccan/discord-backend/internal/auth"
	"github.com/miraccan/discord-backend/internal/config"
	"github.com/miraccan/discord-backend/internal/router"
	"github.com/miraccan/discord-backend/internal/signaling"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	hub := signaling.NewHub()
	go hub.Run()

	issuer := auth.NewIssuer(cfg.JWTSecret, cfg.JWTTTLMinutes)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(cfg, hub, issuer),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("signaling server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
}
