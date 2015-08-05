package types

import (
	"fmt"

	"github.com/Rugvip/bullettime/utils"
)

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

type StreamToken struct {
	MessageIndex  int64
	PresenceIndex int64
	TypingIndex   int64
}

type TokenParseError string

func (e TokenParseError) Error() string {
	return "failed to parse token: " + string(e)
}

func (t StreamToken) String() string {
	return fmt.Sprintf("s%d_%d_%d", t.MessageIndex, t.PresenceIndex, t.TypingIndex)
}

func ParseStreamToken(str string) (StreamToken, error) {
	var message, presence, typing int64
	count, err := fmt.Sscanf(str, "s%d_%d_%d", &message, &presence, &typing)
	if err != nil {
		return StreamToken{}, TokenParseError(err.Error())
	}
	if count != 3 {
		return StreamToken{}, TokenParseError("token does not match format")
	}
	return StreamToken{message, presence, typing}, nil
}

func (t *StreamToken) UnmarshalJSON(bytes []byte) (err error) {
	*t, err = ParseStreamToken(utils.StripQuotes(string(bytes)))
	return
}

func (t StreamToken) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}
