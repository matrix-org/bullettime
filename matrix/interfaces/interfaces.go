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

package interfaces

import (
	"fmt"

	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/types"
)

type RoomService interface {
	CreateRoom(
		hostname string,
		creator ct.UserId,
		desc *types.RoomDescription,
	) (ct.RoomId, *ct.Alias, ct.Error)
	RoomExists(room ct.RoomId, caller ct.UserId) ct.Error
	LookupAlias(alias ct.Alias) (ct.RoomId, ct.Error)
	AddMessage(
		room ct.RoomId,
		caller ct.UserId,
		content ct.TypedContent,
	) (*types.Message, ct.Error)
	State(
		room ct.RoomId,
		caller ct.UserId,
		eventType, stateKey string,
	) (*types.State, ct.Error)
	SetState(
		room ct.RoomId,
		caller ct.UserId,
		content ct.TypedContent,
		stateKey string,
	) (*types.State, ct.Error)
}

type SyncService interface {
	FullSync(user ct.UserId, limit uint) (*types.InitialSync, ct.Error)
	RoomSync(user ct.UserId, room ct.RoomId, limit uint) (*types.RoomInitialSync, ct.Error)
}

type UserService interface {
	CreateUser(ct.UserId) ct.Error
	UserExists(user, caller ct.UserId) (bool, ct.Error)
	VerifyPassword(user ct.UserId, password string) (bool, ct.Error)
	SetPassword(user, caller ct.UserId, password string) ct.Error
}

type ProfileService interface {
	Profile(user, caller ct.UserId) (types.UserProfile, ct.Error)
	UpdateProfile(
		user, caller ct.UserId,
		name, avatarUrl *string,
	) (types.UserProfile, ct.Error)
}

type PresenceService interface {
	Status(user, caller ct.UserId) (types.UserStatus, ct.Error)
	UpdateStatus(
		user, caller ct.UserId,
		presence *types.Presence,
		statusMessage *string,
	) (types.UserStatus, ct.Error)
}

type TokenService interface {
	NewAccessToken(ct.UserId) (Token, ct.Error)
	ParseAccessToken(token string) (Token, ct.Error)
}

type Token interface {
	fmt.Stringer
	UserId() ct.UserId
}

type EventService interface {
	Event(caller ct.UserId, eventId ct.EventId) (ct.Event, ct.Error)
	Range(
		caller ct.UserId,
		from, to *types.StreamToken,
		limit uint,
		cancel chan struct{},
	) (*types.EventStreamRange, ct.Error)
	Messages(
		user ct.UserId,
		room ct.RoomId,
		from, to *types.StreamToken,
		limit uint,
	) (*types.EventStreamRange, ct.Error)
}

type UserStore interface {
	CreateUser(ct.UserId) (exists bool, err ct.Error)
	UserExists(ct.UserId) (exists bool, err ct.Error)
	SetUserPasswordHash(id ct.UserId, hash string) ct.Error
	UserPasswordHash(ct.UserId) (string, ct.Error)
}

type RoomStore interface {
	CreateRoom(id ct.RoomId) (exists bool, err ct.Error)
	RoomExists(ct.RoomId) (bool, ct.Error)
	SetRoomState(roomId ct.RoomId, userId ct.UserId, content ct.TypedContent, stateKey string) (*types.State, ct.Error)
	RoomState(roomId ct.RoomId, eventType, stateKey string) (*types.State, ct.Error)
	EntireRoomState(roomId ct.RoomId) ([]*types.State, ct.Error)
}

type MembershipStore interface {
	AddMember(ct.RoomId, ct.UserId) ct.Error
	RemoveMember(ct.RoomId, ct.UserId) ct.Error
	Rooms(ct.UserId) ([]ct.RoomId, ct.Error)
	Users(ct.RoomId) ([]ct.UserId, ct.Error)
	Peers(ct.UserId) (map[ct.UserId]struct{}, ct.Error)
}

type AsyncEventSink interface {
	Send(userIds []ct.UserId, event ct.IndexedEvent) ct.Error
}

type AsyncEventSource interface {
	Listen(user ct.UserId, cancel chan struct{}) (chan ct.IndexedEvent, ct.Error)
}

type IndexedEventSource interface {
	Max() uint64
	Range(
		user *ct.UserId,
		userSet map[ct.UserId]struct{},
		roomSet map[ct.RoomId]struct{},
		from, to uint64,
		limit uint,
	) ([]ct.IndexedEvent, ct.Error)
}

type EventSink interface {
	Send(event ct.Event) (uint64, ct.Error)
}

type EventProvider interface {
	Event(ct.UserId, ct.EventId) (ct.Event, ct.Error)
}

type ProfileEventSink interface {
	SetUserProfile(ct.UserId, types.UserProfile) (ct.IndexedEvent, ct.Error)
}

type PresenceEventSink interface {
	SetUserStatus(ct.UserId, types.UserStatus) (ct.IndexedEvent, ct.Error)
}

type ProfileProvider interface {
	Profile(ct.UserId) (types.UserProfile, ct.Error)
}

type PresenceProvider interface {
	Status(ct.UserId) (types.UserStatus, ct.Error)
}

type TypingEventSink interface {
	SetTyping(room ct.RoomId, user ct.UserId, typing bool) ct.Error
}

type TypingProvider interface {
	Typing(room ct.RoomId) ([]ct.UserId, ct.Error)
}

type EventStream interface {
	EventSink
	EventProvider
	IndexedEventSource
}

type PresenceStream interface {
	ProfileEventSink
	PresenceEventSink
	ProfileProvider
	PresenceProvider
	IndexedEventSource
}

type TypingStream interface {
	TypingEventSink
	TypingProvider
	IndexedEventSource
}
