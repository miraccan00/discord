package signaling

// Room holds the set of clients currently joined to one voice channel. All
// access happens from the Hub's single goroutine, so no locking is required.
type Room struct {
	id      string
	members map[string]*Client
}

func newRoom(id string) *Room {
	return &Room{id: id, members: make(map[string]*Client)}
}

// add inserts a client into the room.
func (r *Room) add(c *Client) {
	r.members[c.id] = c
}

// remove deletes a client from the room and reports whether the room is now
// empty (so the hub can reclaim it).
func (r *Room) remove(id string) (empty bool) {
	delete(r.members, id)
	return len(r.members) == 0
}

// peersExcept returns the PeerInfo of every member other than the given id.
func (r *Room) peersExcept(id string) []PeerInfo {
	peers := make([]PeerInfo, 0, len(r.members))
	for mid, c := range r.members {
		if mid == id {
			continue
		}
		peers = append(peers, c.info())
	}
	return peers
}

// broadcast queues a message to every member except exceptID.
func (r *Room) broadcast(data []byte, exceptID string) {
	for mid, c := range r.members {
		if mid == exceptID {
			continue
		}
		c.enqueue(data)
	}
}
