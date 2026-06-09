// Package integration exercises the full HTTP + WebSocket signaling stack over a
// real loopback server, complementing the unit tests in internal/signaling.
package integration

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/miraccan/discord-backend/internal/auth"
	"github.com/miraccan/discord-backend/internal/config"
	"github.com/miraccan/discord-backend/internal/router"
	"github.com/miraccan/discord-backend/internal/signaling"
)

func dial(t *testing.T, baseURL, token string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	wsURL := strings.Replace(baseURL, "http://", "ws://", 1) + "/ws?token=" + token
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func read(t *testing.T, conn *websocket.Conn) signaling.Message {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var msg signaling.Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	return msg
}

func write(t *testing.T, conn *websocket.Conn, msg signaling.Message) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestSignalingFlow drives two real WebSocket clients through join and a directed
// offer relay, asserting the wire behaviour end-to-end.
func TestSignalingFlow(t *testing.T) {
	// Arrange: a full server with origin checking disabled for the test client.
	cfg := config.Config{JWTSecret: "test", JWTTTLMinutes: 60, AllowedOrigins: []string{"*"}}
	hub := signaling.NewHub()
	go hub.Run()
	issuer := auth.NewIssuer(cfg.JWTSecret, cfg.JWTTTLMinutes)
	srv := httptest.NewServer(router.New(cfg, hub, issuer))
	defer srv.Close()

	aliceTok, _ := issuer.Issue("alice")
	bobTok, _ := issuer.Issue("bob")

	// Act + Assert: alice joins an empty room.
	alice := dial(t, srv.URL, aliceTok)
	defer alice.Close(websocket.StatusNormalClosure, "")
	write(t, alice, signaling.Message{Type: signaling.TypeJoin})
	if m := read(t, alice); m.Type != signaling.TypeRoomState {
		t.Fatalf("alice first message = %q, want room-state", m.Type)
	}

	// bob joins; alice should be told a peer joined.
	bob := dial(t, srv.URL, bobTok)
	defer bob.Close(websocket.StatusNormalClosure, "")
	write(t, bob, signaling.Message{Type: signaling.TypeJoin})

	bobState := read(t, bob)
	if bobState.Type != signaling.TypeRoomState {
		t.Fatalf("bob first message = %q, want room-state", bobState.Type)
	}
	if m := read(t, alice); m.Type != signaling.TypePeerJoined {
		t.Fatalf("alice = %q, want peer-joined", m.Type)
	}

	// Directed offer from alice must reach bob, stamped with the sender id.
	var state signaling.RoomStatePayload
	if err := json.Unmarshal(bobState.Payload, &state); err != nil {
		t.Fatalf("decode bob room-state: %v", err)
	}
	aliceID := state.Peers[0].ID
	bobID := state.SelfID
	write(t, alice, signaling.Message{
		Type:    signaling.TypeOffer,
		To:      bobID,
		Payload: []byte(`{"sdp":{"type":"offer","sdp":"x"}}`),
	})
	got := read(t, bob)
	if got.Type != signaling.TypeOffer {
		t.Fatalf("bob = %q, want offer", got.Type)
	}
	if got.From != aliceID {
		t.Fatalf("offer from = %q, want alice id %q", got.From, aliceID)
	}
}
