package signaling

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

// sendBuffer is the per-client outbound queue depth. A client that cannot keep
// up (queue full) is disconnected to protect the hub from slow consumers.
const sendBuffer = 32

// pingInterval keeps idle WebSocket connections alive through proxies.
const pingInterval = 30 * time.Second

// Client represents a single WebSocket connection owned by one user.
type Client struct {
	id       string
	username string
	roomID   string

	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// State mirrored to peers via mute-state.
	muted    bool
	deafened bool
}

// Serve attaches a freshly accepted WebSocket connection to the hub and blocks
// until it disconnects. The caller must authenticate the user beforehand.
func (h *Hub) Serve(ctx context.Context, conn *websocket.Conn, username string) {
	c := newClient(h, conn, username)
	go c.writePump(ctx)
	c.readPump(ctx)
}

// newClient wraps an accepted WebSocket connection.
func newClient(hub *Hub, conn *websocket.Conn, username string) *Client {
	return &Client{
		id:       uuid.NewString(),
		username: username,
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBuffer),
	}
}

// info returns the peer-facing view of this client.
func (c *Client) info() PeerInfo {
	return PeerInfo{ID: c.id, Name: c.username, Muted: c.muted, Deafened: c.deafened}
}

// enqueue queues a pre-serialized message for delivery. It returns false if the
// client's buffer is full, signalling that it should be dropped.
func (c *Client) enqueue(data []byte) bool {
	select {
	case c.send <- data:
		return true
	default:
		return false
	}
}

// readPump reads inbound messages until the connection closes, forwarding each
// to the hub. It always unregisters the client on exit.
func (c *Client) readPump(ctx context.Context) {
	defer func() { c.hub.unregister <- c }()
	for {
		var msg Message
		if err := wsjson.Read(ctx, c.conn, &msg); err != nil {
			return
		}
		// The server is authoritative over the sender identity.
		msg.From = c.id
		select {
		case c.hub.inbound <- inbound{client: c, msg: msg}:
		case <-ctx.Done():
			return
		}
	}
}

// writePump drains the send queue to the socket and emits periodic pings.
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				_ = c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.Ping(ctx); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// sendTyped serializes and queues a typed server message to this client.
func (c *Client) sendTyped(msgType string, payload any) {
	msg := Message{Type: msgType}
	if payload != nil {
		msg.Payload = mustMarshal(payload)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.enqueue(data)
}
