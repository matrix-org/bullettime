package types

type Pagination struct {
	Start string        `json:"start"`
	End   string        `json:"end"`
	Chunk []interface{} `json:"chunk"`
}

type InitialSync struct {
	End      string          `json:"end"`
	Presence []PresenceEvent `json:"presence"`
	Rooms    []RoomSummary   `json:"rooms"`
}

type RoomSummary struct {
	Membership Membership `json:"membership"`
	RoomId     RoomId     `json:"room_id"`
	Messages   []Event    `json:"messages"`
	State      []State    `json:"state"`
}

type RoomInitialSync struct {
	RoomSummary
	Presence PresenceEvent `json:"presence"`
}
