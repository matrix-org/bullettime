package interfaces

import (
	"fmt"

	"github.com/Rugvip/bullettime/types"
)

type RoomService interface {
	CreateRoom(
		hostname string,
		creator types.UserId,
		desc *types.RoomDescription,
	) (types.RoomId, *types.Alias, types.Error)
	RoomExists(room types.RoomId, caller types.UserId) types.Error
	AddMessage(
		room types.RoomId,
		caller types.UserId,
		content types.TypedContent,
	) (*types.Message, types.Error)
	State(
		room types.RoomId,
		caller types.UserId,
		eventType, stateKey string,
	) (*types.State, types.Error)
	SetState(
		room types.RoomId,
		caller types.UserId,
		content types.TypedContent,
		stateKey string,
	) (*types.State, types.Error)
}

type SyncService interface {
	FullSync(user types.UserId, limit uint) (*types.InitialSync, types.Error)
	RoomSync(user types.UserId, room types.RoomId, limit uint) (*types.RoomInitialSync, types.Error)
}

type UserService interface {
	CreateUser(types.UserId) types.Error
	UserExists(user, caller types.UserId) types.Error
	VerifyPassword(user types.UserId, password string) types.Error
	SetPassword(user, caller types.UserId, password string) types.Error
}

type ProfileService interface {
	Profile(user, caller types.UserId) (types.UserProfile, types.Error)
	UpdateProfile(
		user, caller types.UserId,
		name, avatarUrl *string,
	) (types.UserProfile, types.Error)
}

type PresenceService interface {
	Status(user, caller types.UserId) (types.UserStatus, types.Error)
	UpdateStatus(
		user, caller types.UserId,
		presence *types.Presence,
		statusMessage *string,
	) (types.UserStatus, types.Error)
}

type TokenService interface {
	NewAccessToken(types.UserId) (Token, types.Error)
	ParseAccessToken(token string) (Token, types.Error)
}

type Token interface {
	fmt.Stringer
	UserId() types.UserId
}

type EventService interface {
	Event(caller types.UserId, eventId types.EventId) (types.Event, types.Error)
	Range(
		caller types.UserId,
		from, to *types.StreamToken,
		limit uint,
		cancel chan struct{},
	) (*types.EventStreamChunk, types.Error)
}

type UserStore interface {
	CreateUser(types.UserId) types.Error
	UserExists(types.UserId) (bool, types.Error)
	SetUserPasswordHash(id types.UserId, hash string) types.Error
	UserPasswordHash(types.UserId) (string, types.Error)
}

type RoomStore interface {
	CreateRoom(domain string) (types.RoomId, types.Error)
	RoomExists(types.RoomId) (bool, types.Error)
	AddRoomMessage(types.RoomId, types.UserId, types.TypedContent) (*types.Message, types.Error)
	SetRoomState(roomId types.RoomId, userId types.UserId, content types.TypedContent, stateKey string) (*types.State, types.Error)
	RoomState(roomId types.RoomId, eventType, stateKey string) (*types.State, types.Error)
	EntireRoomState(roomId types.RoomId) ([]*types.State, types.Error)
}

type AliasStore interface {
	Reserve(alias types.Alias) types.Error
	Claim(alias types.Alias, roomId types.RoomId) types.Error
	AddAlias(types.Alias, types.RoomId) types.Error
	RemoveAlias(types.Alias, types.RoomId) types.Error
	Aliases(types.RoomId) ([]types.Alias, types.Error)
	Room(types.Alias) (*types.RoomId, types.Error)
}

type MembershipStore interface {
	AddMember(types.RoomId, types.UserId) types.Error
	RemoveMember(types.RoomId, types.UserId) types.Error
	Rooms(types.UserId) ([]types.RoomId, types.Error)
	Users(types.RoomId) ([]types.UserId, types.Error)
	Peers(types.UserId) (map[types.UserId]struct{}, types.Error)
}

type AsyncEventSink interface {
	Send(userIds []types.UserId, event types.IndexedEvent) types.Error
}

type AsyncEventSource interface {
	Listen(user types.UserId, cancel chan struct{}) (chan types.IndexedEvent, types.Error)
}

type IndexedEventSource interface {
	Max() uint64
	Range(
		user types.UserId,
		userSet map[types.UserId]struct{},
		roomSet map[types.RoomId]struct{},
		from, to uint64,
		limit uint,
	) ([]types.IndexedEvent, types.Error)
}

type EventSink interface {
	Send(event types.Event) (uint64, types.Error)
}

type EventProvider interface {
	Event(types.UserId, types.EventId) (types.Event, types.Error)
}

type ProfileEventSink interface {
	SetUserProfile(types.UserId, types.UserProfile) (types.IndexedEvent, types.Error)
}

type PresenceEventSink interface {
	SetUserStatus(types.UserId, types.UserStatus) (types.IndexedEvent, types.Error)
}

type ProfileProvider interface {
	Profile(types.UserId) (types.UserProfile, types.Error)
}

type PresenceProvider interface {
	Status(types.UserId) (types.UserStatus, types.Error)
}

type TypingEventSink interface {
	SetTyping(room types.RoomId, user types.UserId, typing bool) types.Error
}

type TypingProvider interface {
	Typing(room types.RoomId) ([]types.UserId, types.Error)
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
