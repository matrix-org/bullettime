package types

import (
	"encoding/json"
	"strconv"
	"time"
)

const (
	EventTypeCreate      string = "m.room.create"
	EventTypeName               = "m.room.name"
	EventTypeTopic              = "m.room.topic"
	EventTypeAliases            = "m.room.aliases"
	EventTypeJoinRules          = "m.room.join_rules"
	EventTypeMembership         = "m.room.member"
	EventTypePowerLevels        = "m.room.power_levels"
)

type TypedContent interface {
	EventType() string
}

type Content interface{}

type Timestamp struct {
	time.Time
}

type Event struct {
	EventId   EventId   `json:"event_id"`
	RoomId    RoomId    `json:"room_id"`
	UserId    UserId    `json:"user_id"`
	EventType string    `json:"type"`
	Timestamp Timestamp `json:"origin_server_ts"`
	Content   Content   `json:"content"`
}

type OldState State

type State struct {
	Event
	StateKey string    `json:"state_key"`
	OldState *OldState `json:"prev_content"`
}

func (e *OldState) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	return json.Marshal(e.Content)
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	ms := ts.UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(ms, 10)), nil
}

type GenericContent struct {
	Content   map[string]interface{}
	eventType string
}

func (c *GenericContent) EventType() string {
	return c.eventType
}

func (c *GenericContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Content)
}

type TestContent struct {
	Name string `json:"name"`
}

type MembershipEventContent struct {
	*UserProfile
	Membership Membership `json:"membership"`
}

func (c *MembershipEventContent) EventType() string {
	return EventTypeMembership
}

type CreateEventContent struct {
	Creator UserId `json:"creator"`
}

func (c *CreateEventContent) EventType() string {
	return EventTypeCreate
}

type NameEventContent struct {
	Name string `json:"name"`
}

func (c *NameEventContent) EventType() string {
	return EventTypeName
}

type TopicEventContent struct {
	Topic string `json:"topic"`
}

func (c *TopicEventContent) EventType() string {
	return EventTypeTopic
}

type AliasesEventContent struct {
	Aliases []Alias `json:"aliases"`
}

func (c *AliasesEventContent) EventType() string {
	return EventTypeAliases
}

func DefaultPowerLevels(creator UserId) *PowerLevelsEventContent {
	powerLevels := new(PowerLevelsEventContent)
	powerLevels.Ban = 50
	powerLevels.Kick = 50
	powerLevels.Redact = 50
	powerLevels.CreateState = 50
	powerLevels.EventDefault = 0
	powerLevels.Users = map[UserId]int{
		creator: 100,
	}
	powerLevels.Events = map[string]int{
		"m.room.name":         100,
		"m.room.power_levels": 100,
	}
	return powerLevels
}

type PowerLevelsEventContent struct {
	Ban          int            `json:"ban"`
	Kick         int            `json:"kick"`
	Redact       int            `json:"redact"`
	UserDefault  int            `json:"users_default"`
	CreateState  int            `json:"state_default"`
	EventDefault int            `json:"events_default"`
	Users        map[UserId]int `json:"users"`
	Events       map[string]int `json:"events"`
}

func (c *PowerLevelsEventContent) EventType() string {
	return EventTypePowerLevels
}

type JoinRulesEventContent struct {
	JoinRule JoinRule `json:"join_rule"`
}

func (c *JoinRulesEventContent) EventType() string {
	return EventTypeJoinRules
}
