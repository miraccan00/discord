package signaling

import (
	"encoding/json"
	"testing"
)

// newTestClient builds a client with a buffered send channel and no real
// socket, so hub routing logic can be exercised directly and synchronously.
func newTestClient(h *Hub, id, name string) *Client {
	return &Client{id: id, username: name, hub: h, send: make(chan []byte, 16)}
}

// drain returns all queued messages for a client, decoded.
func drain(t *testing.T, c *Client) []Message {
	t.Helper()
	var out []Message
	for {
		select {
		case data := <-c.send:
			var m Message
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("unmarshal queued message: %v", err)
			}
			out = append(out, m)
		default:
			return out
		}
	}
}

func TestJoinBroadcastsPeerJoinedAndRoomState(t *testing.T) {
	// Arrange
	hub := NewHub()
	alice := newTestClient(hub, "a", "alice")
	bob := newTestClient(hub, "b", "bob")

	// Act: alice joins first, then bob joins.
	hub.handleInbound(inbound{client: alice, msg: Message{Type: TypeJoin}})
	hub.handleInbound(inbound{client: bob, msg: Message{Type: TypeJoin}})

	// Assert: alice received an (empty) room-state then a peer-joined for bob.
	aliceMsgs := drain(t, alice)
	if len(aliceMsgs) != 2 {
		t.Fatalf("alice got %d messages, want 2: %+v", len(aliceMsgs), aliceMsgs)
	}
	if aliceMsgs[0].Type != TypeRoomState {
		t.Fatalf("alice first message = %q, want %q", aliceMsgs[0].Type, TypeRoomState)
	}
	if aliceMsgs[1].Type != TypePeerJoined {
		t.Fatalf("alice second message = %q, want %q", aliceMsgs[1].Type, TypePeerJoined)
	}

	// Assert: bob's room-state lists alice as an existing peer.
	bobMsgs := drain(t, bob)
	if len(bobMsgs) != 1 || bobMsgs[0].Type != TypeRoomState {
		t.Fatalf("bob messages = %+v, want a single room-state", bobMsgs)
	}
	var state RoomStatePayload
	if err := json.Unmarshal(bobMsgs[0].Payload, &state); err != nil {
		t.Fatalf("decode room-state: %v", err)
	}
	if state.SelfID != "b" || len(state.Peers) != 1 || state.Peers[0].ID != "a" {
		t.Fatalf("room-state = %+v, want self=b with peer a", state)
	}
}

func TestRelayDirectedOnlyReachesTarget(t *testing.T) {
	// Arrange: three clients in one room.
	hub := NewHub()
	alice := newTestClient(hub, "a", "alice")
	bob := newTestClient(hub, "b", "bob")
	carol := newTestClient(hub, "c", "carol")
	for _, c := range []*Client{alice, bob, carol} {
		hub.handleInbound(inbound{client: c, msg: Message{Type: TypeJoin}})
	}
	// Clear join-related traffic.
	drain(t, alice)
	drain(t, bob)
	drain(t, carol)

	// Act: alice sends an offer addressed to bob.
	offer := Message{Type: TypeOffer, From: "a", To: "b", Payload: json.RawMessage(`{"sdp":"x"}`)}
	hub.handleInbound(inbound{client: alice, msg: offer})

	// Assert: only bob receives it, stamped with from=a.
	if got := drain(t, bob); len(got) != 1 || got[0].Type != TypeOffer || got[0].From != "a" {
		t.Fatalf("bob messages = %+v, want one offer from a", got)
	}
	if got := drain(t, carol); len(got) != 0 {
		t.Fatalf("carol should receive nothing, got %+v", got)
	}
	if got := drain(t, alice); len(got) != 0 {
		t.Fatalf("alice (sender) should receive nothing, got %+v", got)
	}
}

func TestChatBroadcastsToEveryone(t *testing.T) {
	// Arrange
	hub := NewHub()
	alice := newTestClient(hub, "a", "alice")
	bob := newTestClient(hub, "b", "bob")
	hub.handleInbound(inbound{client: alice, msg: Message{Type: TypeJoin}})
	hub.handleInbound(inbound{client: bob, msg: Message{Type: TypeJoin}})
	drain(t, alice)
	drain(t, bob)

	// Act
	chat := Message{Type: TypeChatMessage, From: "a", Payload: json.RawMessage(`{"text":"hi"}`)}
	hub.handleInbound(inbound{client: alice, msg: chat})

	// Assert: both the sender and the peer receive the canonical chat message.
	for _, c := range []*Client{alice, bob} {
		got := drain(t, c)
		if len(got) != 1 || got[0].Type != TypeChatMessage {
			t.Fatalf("%s messages = %+v, want one chat-message", c.username, got)
		}
		var p ChatPayload
		if err := json.Unmarshal(got[0].Payload, &p); err != nil {
			t.Fatalf("decode chat payload: %v", err)
		}
		if p.Text != "hi" || p.Name != "alice" {
			t.Fatalf("chat payload = %+v, want text=hi name=alice", p)
		}
	}
}

func TestLeaveBroadcastsPeerLeft(t *testing.T) {
	// Arrange
	hub := NewHub()
	alice := newTestClient(hub, "a", "alice")
	bob := newTestClient(hub, "b", "bob")
	hub.handleInbound(inbound{client: alice, msg: Message{Type: TypeJoin}})
	hub.handleInbound(inbound{client: bob, msg: Message{Type: TypeJoin}})
	drain(t, alice)
	drain(t, bob)

	// Act: bob disconnects.
	hub.handleLeave(bob)

	// Assert: alice is told bob left.
	got := drain(t, alice)
	if len(got) != 1 || got[0].Type != TypePeerLeft {
		t.Fatalf("alice messages = %+v, want one peer-left", got)
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(got[0].Payload, &p); err != nil {
		t.Fatalf("decode peer-left: %v", err)
	}
	if p.ID != "b" {
		t.Fatalf("peer-left id = %q, want b", p.ID)
	}
}
