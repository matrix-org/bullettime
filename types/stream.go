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
	End      StreamToken   `json:"end"`
	Presence []Event       `json:"presence"`
	Rooms    []RoomSummary `json:"rooms"`
}

type RoomSummary struct {
	Membership Membership `json:"membership"`
	RoomId     RoomId     `json:"room_id"`
	Messages   []Event    `json:"messages"`
	State      []*State   `json:"state"`
	Visibility Visibility `json:"visibility"`
}

type RoomInitialSync struct {
	RoomSummary
	Presence []Event `json:"presence"`
}

type EventStreamChunk struct {
	Events []Event     `json:"chunk"`
	Start  StreamToken `json:"start"`
	End    StreamToken `json:"end"`
}

type StreamToken struct {
	MessageIndex  uint64
	PresenceIndex uint64
	TypingIndex   uint64
}

type TokenParseError string

func (e TokenParseError) Error() string {
	return "failed to parse token: " + string(e)
}

func (t StreamToken) String() string {
	return fmt.Sprintf("s%d_%d_%d", t.MessageIndex, t.PresenceIndex, t.TypingIndex)
}

func NewEventStreamChunk(events []Event, start StreamToken, end StreamToken) *EventStreamChunk {
	return &EventStreamChunk{
		Events: events,
		Start:  start,
		End:    end,
	}
}

func NewStreamToken(messageIndex, presenceIndex, typingIndex uint64) StreamToken {
	return StreamToken{
		MessageIndex:  messageIndex,
		PresenceIndex: presenceIndex,
		TypingIndex:   typingIndex,
	}
}

func ParseStreamToken(str string) (StreamToken, error) {
	var message, presence, typing uint64
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
