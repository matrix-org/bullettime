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
	) (ct.RoomId, *ct.Alias, types.Error)
	RoomExists(room ct.RoomId, caller ct.UserId) types.Error
	LookupAlias(alias ct.Alias) (ct.RoomId, types.Error)
	AddMessage(
		room ct.RoomId,
		caller ct.UserId,
		content ct.TypedContent,
	) (*types.Message, types.Error)
	State(
		room ct.RoomId,
		caller ct.UserId,
		eventType, stateKey string,
	) (*types.State, types.Error)
	SetState(
		room ct.RoomId,
		caller ct.UserId,
		content ct.TypedContent,
		stateKey string,
	) (*types.State, types.Error)
}

type SyncService interface {
	FullSync(user ct.UserId, limit uint) (*types.InitialSync, types.Error)
	RoomSync(user ct.UserId, room ct.RoomId, limit uint) (*types.RoomInitialSync, types.Error)
}

type UserService interface {
	CreateUser(ct.UserId) types.Error
	UserExists(user, caller ct.UserId) (bool, types.Error)
	VerifyPassword(user ct.UserId, password string) (bool, types.Error)
	SetPassword(user, caller ct.UserId, password string) types.Error
}

type ProfileService interface {
	Profile(user, caller ct.UserId) (types.UserProfile, types.Error)
	UpdateProfile(
		user, caller ct.UserId,
		name, avatarUrl *string,
	) (types.UserProfile, types.Error)
}

type PresenceService interface {
	Status(user, caller ct.UserId) (types.UserStatus, types.Error)
	UpdateStatus(
		user, caller ct.UserId,
		presence *types.Presence,
		statusMessage *string,
	) (types.UserStatus, types.Error)
}

type TokenService interface {
	NewAccessToken(ct.UserId) (Token, types.Error)
	ParseAccessToken(token string) (Token, types.Error)
}

type Token interface {
	fmt.Stringer
	UserId() ct.UserId
}

type EventService interface {
	Event(caller ct.UserId, eventId ct.EventId) (ct.Event, types.Error)
	Range(
		caller ct.UserId,
		from, to *types.StreamToken,
		limit uint,
		cancel chan struct{},
	) (*types.EventStreamRange, types.Error)
	Messages(
		user ct.UserId,
		room ct.RoomId,
		from, to *types.StreamToken,
		limit uint,
	) (*types.EventStreamRange, types.Error)
}

type UserStore interface {
	CreateUser(ct.UserId) (exists bool, err types.Error)
	UserExists(ct.UserId) (exists bool, err types.Error)
	SetUserPasswordHash(id ct.UserId, hash string) types.Error
	UserPasswordHash(ct.UserId) (string, types.Error)
}

type RoomStore interface {
	CreateRoom(id ct.RoomId) (exists bool, err types.Error)
	RoomExists(ct.RoomId) (bool, types.Error)
	SetRoomState(roomId ct.RoomId, userId ct.UserId, content ct.TypedContent, stateKey string) (*types.State, types.Error)
	RoomState(roomId ct.RoomId, eventType, stateKey string) (*types.State, types.Error)
	EntireRoomState(roomId ct.RoomId) ([]*types.State, types.Error)
}

type AliasStore interface {
	AddAlias(ct.Alias, ct.RoomId) types.Error
	RemoveAlias(ct.Alias, ct.RoomId) types.Error
	Aliases(ct.RoomId) ([]ct.Alias, types.Error)
	Room(ct.Alias) (*ct.RoomId, types.Error)
}

type MembershipStore interface {
	AddMember(ct.RoomId, ct.UserId) types.Error
	RemoveMember(ct.RoomId, ct.UserId) types.Error
	Rooms(ct.UserId) ([]ct.RoomId, types.Error)
	Users(ct.RoomId) ([]ct.UserId, types.Error)
	Peers(ct.UserId) (map[ct.UserId]struct{}, types.Error)
}

type AsyncEventSink interface {
	Send(userIds []ct.UserId, event ct.IndexedEvent) types.Error
}

type AsyncEventSource interface {
	Listen(user ct.UserId, cancel chan struct{}) (chan ct.IndexedEvent, types.Error)
}

type IndexedEventSource interface {
	Max() uint64
	Range(
		user *ct.UserId,
		userSet map[ct.UserId]struct{},
		roomSet map[ct.RoomId]struct{},
		from, to uint64,
		limit uint,
	) ([]ct.IndexedEvent, types.Error)
}

type EventSink interface {
	Send(event ct.Event) (uint64, types.Error)
}

type EventProvider interface {
	Event(ct.UserId, ct.EventId) (ct.Event, types.Error)
}

type ProfileEventSink interface {
	SetUserProfile(ct.UserId, types.UserProfile) (ct.IndexedEvent, types.Error)
}

type PresenceEventSink interface {
	SetUserStatus(ct.UserId, types.UserStatus) (ct.IndexedEvent, types.Error)
}

type ProfileProvider interface {
	Profile(ct.UserId) (types.UserProfile, types.Error)
}

type PresenceProvider interface {
	Status(ct.UserId) (types.UserStatus, types.Error)
}

type TypingEventSink interface {
	SetTyping(room ct.RoomId, user ct.UserId, typing bool) types.Error
}

type TypingProvider interface {
	Typing(room ct.RoomId) ([]ct.UserId, types.Error)
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
