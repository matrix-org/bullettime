package types

import (
	"strconv"
	"time"
)

type EventType string

const (
	EventTypeCreate      EventType = "m.room.create"
	EventTypeName                  = "m.room.name"
	EventTypeTopic                 = "m.room.topic"
	EventTypeMember                = "m.room.member"
	EventTypeAliases               = "m.room.aliases"
	EventTypeJoinRules             = "m.room.join_rules"
	EventTypePowerLevels           = "m.room.power_levels"
)

type Content interface{}

type TimeStamp struct {
	time.Time
}

type Event struct {
	Id        EventId   `json:"event_id"`
	RoomId    RoomId    `json:"room_id"`
	EventType EventType `json:"type"`
	TimeStamp TimeStamp `json:"origin_server_ts"`
	Content   Content   `json:"content"`
}

type State struct {
	Event
	StateKey   string  `json:"state_key"`
	OldContent Content `json:"prev_content"`
}

func (ts TimeStamp) MarshalJSON() ([]byte, error) {
	ms := ts.UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(ms, 10)), nil
}

func (ts *TimeStamp) UnmarshalJSON(data []byte) error {
	ms, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}

	ts.Time = time.Unix(0, ms*int64(time.Millisecond))
	return nil
}

type TestContent struct {
	Name string `json:"name"`
}

type MembershipContent struct {
	Membership  string `json:"membership"`
	DisplayName string `json:"displayname"`
	AvatarUrl   string `json:"avatar_url"`
}
