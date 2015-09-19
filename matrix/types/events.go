// Copyright 2015  Ericsson AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/json"

	ct "github.com/matrix-org/bullettime/core/types"
)

const (
	EventTypeCreate      = "m.room.create"
	EventTypeName        = "m.room.name"
	EventTypeTopic       = "m.room.topic"
	EventTypeAliases     = "m.room.aliases"
	EventTypeJoinRules   = "m.room.join_rules"
	EventTypeMembership  = "m.room.member"
	EventTypePowerLevels = "m.room.power_levels"
	EventTypeTyping      = "m.typing"
	EventTypePresence    = "m.presence"
)

type BaseEvent struct {
	EventType string `json:"type"`
}

func (e *BaseEvent) GetEventType() string {
	return e.EventType
}

type Message struct {
	BaseEvent
	Content   ct.Content   `json:"content"`
	EventId   ct.EventId   `json:"event_id"`
	RoomId    ct.RoomId    `json:"room_id"`
	UserId    ct.UserId    `json:"user_id"`
	Timestamp ct.Timestamp `json:"origin_server_ts"`
}

func (e *Message) GetContent() interface{} {
	return e.Content
}

func (e *Message) GetEventId() *ct.EventId {
	return &e.EventId
}

func (e *Message) GetRoomId() *ct.RoomId {
	return &e.RoomId
}

func (e *Message) GetUserId() *ct.UserId {
	return &e.UserId
}

func (e *Message) GetEventKey() ct.Id {
	return ct.Id(e.EventId)
}

type PresenceEvent struct {
	BaseEvent
	Content User `json:"content"`
}

func (e *PresenceEvent) GetEventType() string {
	return EventTypePresence
}

func (e *PresenceEvent) GetContent() interface{} {
	return e.Content
}

func (e *PresenceEvent) GetRoomId() *ct.RoomId {
	return nil
}

func (e *PresenceEvent) GetUserId() *ct.UserId {
	return &e.Content.UserId
}

func (e *PresenceEvent) GetEventKey() ct.Id {
	return ct.Id(e.Content.UserId)
}

type TypingUsers struct {
	UserIds []ct.UserId `json:"user_ids"`
}

type TypingEvent struct {
	BaseEvent
	Content TypingUsers `json:"content"`
	RoomId  ct.RoomId   `json:"room_id"`
}

func (e *TypingEvent) GetEventType() string {
	return EventTypeTyping
}

func (e *TypingEvent) GetContent() interface{} {
	return e.Content
}

func (e *TypingEvent) GetRoomId() *ct.RoomId {
	return &e.RoomId
}

func (e *TypingEvent) GetUserId() *ct.UserId {
	return nil
}

func (e *TypingEvent) GetEventKey() ct.Id {
	return ct.Id(e.RoomId)
}

type OldState State

type State struct {
	Message
	StateKey string    `json:"state_key"`
	OldState *OldState `json:"prev_content"`
}

func (e *OldState) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	return json.Marshal(e.Content)
}

func NewGenericContent(content map[string]interface{}, eventType string) *GenericContent {
	return &GenericContent{content, eventType}
}

type GenericContent struct {
	Content   map[string]interface{}
	eventType string
}

func (c *GenericContent) GetEventType() string {
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

func (c *MembershipEventContent) GetEventType() string {
	return EventTypeMembership
}

type CreateEventContent struct {
	Creator ct.UserId `json:"creator"`
}

func (c *CreateEventContent) GetEventType() string {
	return EventTypeCreate
}

type NameEventContent struct {
	Name string `json:"name"`
}

func (c *NameEventContent) GetEventType() string {
	return EventTypeName
}

type TopicEventContent struct {
	Topic string `json:"topic"`
}

func (c *TopicEventContent) GetEventType() string {
	return EventTypeTopic
}

type AliasesEventContent struct {
	Aliases []ct.Alias `json:"aliases"`
}

func (c *AliasesEventContent) GetEventType() string {
	return EventTypeAliases
}

func DefaultPowerLevels(creator ct.UserId) *PowerLevelsEventContent {
	powerLevels := new(PowerLevelsEventContent)
	powerLevels.Ban = 50
	powerLevels.Kick = 50
	powerLevels.Invite = 0
	powerLevels.Redact = 50
	powerLevels.CreateState = 50
	powerLevels.EventDefault = 0
	powerLevels.Users = UserPowerLevelMap{
		creator.String(): 100,
	}
	powerLevels.Events = map[string]int{
		"m.room.name":         100,
		"m.room.power_levels": 100,
	}
	return powerLevels
}

type PowerLevelsEventContent struct {
	Ban          int               `json:"ban"`
	Kick         int               `json:"kick"`
	Invite       int               `json:"invite"`
	Redact       int               `json:"redact"`
	UserDefault  int               `json:"users_default"`
	CreateState  int               `json:"state_default"`
	EventDefault int               `json:"events_default"`
	Users        UserPowerLevelMap `json:"users"`
	Events       map[string]int    `json:"events"`
}

type UserPowerLevelMap map[string]int

func (m *UserPowerLevelMap) UnmarshalJSON(bytes []byte) error {
	userMap := map[string]int{}
	err := json.Unmarshal(bytes, userMap)
	if err != nil {
		return err
	}
	for userId := range userMap {
		_, err := ct.ParseUserId(userId)
		if err != nil {
			return err
		}
	}
	*m = userMap
	return nil
}

func (c *PowerLevelsEventContent) GetEventType() string {
	return EventTypePowerLevels
}

type JoinRulesEventContent struct {
	JoinRule JoinRule `json:"join_rule"`
}

func (c *JoinRulesEventContent) GetEventType() string {
	return EventTypeJoinRules
}
