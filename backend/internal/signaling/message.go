package signaling

import "encoding/json"

// Message types exchanged over the signaling WebSocket. The same envelope is
// used in both directions; the server stamps From authoritatively.
const (
	// Client -> Server.
	TypeJoin         = "join"
	TypeOffer        = "offer"
	TypeAnswer       = "answer"
	TypeICECandidate = "ice-candidate"
	TypeChatMessage  = "chat-message"
	TypeMuteState    = "mute-state"

	// Server -> Client.
	TypeRoomState  = "room-state"
	TypePeerJoined = "peer-joined"
	TypePeerLeft   = "peer-left"
	TypeError      = "error"
)

// Message is the signaling envelope. Payload is left as raw JSON so the hub can
// relay directed messages (offer/answer/ice-candidate) without understanding
// their contents.
type Message struct {
	Type    string          `json:"type"`
	From    string          `json:"from,omitempty"`
	To      string          `json:"to,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// PeerInfo describes a room member as advertised to other clients.
type PeerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Muted    bool   `json:"muted"`
	Deafened bool   `json:"deafened"`
}

// RoomStatePayload is sent to a joining client so it knows who is already
// present and can initiate offers to them.
type RoomStatePayload struct {
	SelfID   string     `json:"selfId"`
	SelfName string     `json:"selfName"`
	Peers    []PeerInfo `json:"peers"`
}

// ChatPayload is the server-broadcast form of a chat message.
type ChatPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
	TS   int64  `json:"ts"`
}

// mustMarshal serializes a payload, panicking only on programmer error
// (unencodable types), which cannot happen for these concrete structs.
func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
