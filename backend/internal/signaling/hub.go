// Package signaling implements the in-memory room hub that relays WebRTC
// signaling messages (offer/answer/ICE) between connected WebSocket clients.
package signaling

import (
	"encoding/json"
	"time"
)

// defaultRoom is used when a client joins without naming a room.
const defaultRoom = "general"

// inbound couples a received message with its originating client so the hub can
// route it.
type inbound struct {
	client *Client
	msg    Message
}

// Hub is the single owner of all room state. Every mutation flows through its
// Run goroutine via channels, which removes the need for mutexes and keeps the
// design race-free under `go test -race`.
type Hub struct {
	rooms map[string]*Room

	unregister chan *Client
	inbound    chan inbound
}

// NewHub constructs an empty hub.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		unregister: make(chan *Client),
		inbound:    make(chan inbound, 256),
	}
}

// Run processes hub events until the program exits. Call it in its own
// goroutine before accepting connections.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.unregister:
			h.handleLeave(c)
		case in := <-h.inbound:
			h.handleInbound(in)
		}
	}
}

func (h *Hub) handleInbound(in inbound) {
	switch in.msg.Type {
	case TypeJoin:
		h.handleJoin(in.client, in.msg)
	case TypeOffer, TypeAnswer, TypeICECandidate:
		h.relayDirected(in.client, in.msg)
	case TypeChatMessage:
		h.handleChat(in.client, in.msg)
	case TypeMuteState:
		h.handleMuteState(in.client, in.msg)
	default:
		in.client.sendTyped(TypeError, map[string]string{
			"code": "unknown_type", "message": "unknown message type: " + in.msg.Type,
		})
	}
}

func (h *Hub) handleJoin(c *Client, msg Message) {
	roomID := defaultRoom
	if len(msg.Payload) > 0 {
		var p struct {
			Room string `json:"room"`
		}
		if err := json.Unmarshal(msg.Payload, &p); err == nil && p.Room != "" {
			roomID = p.Room
		}
	}

	room := h.rooms[roomID]
	if room == nil {
		room = newRoom(roomID)
		h.rooms[roomID] = room
	}
	c.roomID = roomID

	// Tell the joiner who is already present so it can initiate offers.
	c.sendTyped(TypeRoomState, RoomStatePayload{
		SelfID:   c.id,
		SelfName: c.username,
		Peers:    room.peersExcept(c.id),
	})

	// Announce the newcomer to existing members, then add it to the room.
	if data, ok := marshalMessage(Message{Type: TypePeerJoined, Payload: mustMarshal(c.info())}); ok {
		room.broadcast(data, c.id)
	}
	room.add(c)
}

func (h *Hub) handleLeave(c *Client) {
	close(c.send)
	if c.roomID == "" {
		return
	}
	room := h.rooms[c.roomID]
	if room == nil {
		return
	}
	if empty := room.remove(c.id); empty {
		delete(h.rooms, c.roomID)
		return
	}
	if data, ok := marshalMessage(Message{
		Type:    TypePeerLeft,
		Payload: mustMarshal(map[string]string{"id": c.id}),
	}); ok {
		room.broadcast(data, c.id)
	}
}

// relayDirected forwards offer/answer/ice-candidate to a single target peer in
// the same room.
func (h *Hub) relayDirected(c *Client, msg Message) {
	room := h.rooms[c.roomID]
	if room == nil || msg.To == "" {
		return
	}
	target, ok := room.members[msg.To]
	if !ok {
		return
	}
	if data, ok := marshalMessage(msg); ok {
		target.enqueue(data)
	}
}

func (h *Hub) handleChat(c *Client, msg Message) {
	room := h.rooms[c.roomID]
	if room == nil {
		return
	}
	var p struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Payload, &p); err != nil || p.Text == "" {
		return
	}
	out := Message{
		Type: TypeChatMessage,
		From: c.id,
		Payload: mustMarshal(ChatPayload{
			ID:   c.id,
			Name: c.username,
			Text: p.Text,
			TS:   time.Now().UnixMilli(),
		}),
	}
	if data, ok := marshalMessage(out); ok {
		room.broadcast(data, "") // include the sender for an authoritative copy
	}
}

func (h *Hub) handleMuteState(c *Client, msg Message) {
	room := h.rooms[c.roomID]
	if room == nil {
		return
	}
	var p struct {
		Muted    bool `json:"muted"`
		Deafened bool `json:"deafened"`
	}
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return
	}
	c.muted = p.Muted
	c.deafened = p.Deafened
	out := Message{Type: TypeMuteState, From: c.id, Payload: mustMarshal(c.info())}
	if data, ok := marshalMessage(out); ok {
		room.broadcast(data, c.id)
	}
}

// marshalMessage serializes an envelope, reporting failure rather than
// panicking so a single bad message cannot take down the hub.
func marshalMessage(msg Message) ([]byte, bool) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, false
	}
	return data, true
}
