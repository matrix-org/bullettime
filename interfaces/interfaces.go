package interfaces

import (
	"fmt"

	"github.com/Rugvip/bullettime/types"
)

type RoomService interface {
	Room(types.RoomId) (Room, types.Error)
	CreateRoom(hostname string, creator User, desc *types.RoomDescription) (Room, *types.Alias, types.Error)
}

type Room interface {
	Id() types.RoomId
	AddMessage(user User, content types.TypedContent) (*types.Event, types.Error)
	SetState(user User, content types.TypedContent, stateKey string) (*types.State, types.Error)
	State(user User, eventType, stateKey string) (*types.State, types.Error)
}

type UserService interface {
	User(types.UserId) (User, types.Error)
	CreateUser(types.UserId) (User, types.Error)
}

type User interface {
	Id() types.UserId
	VerifyPassword(password string) types.Error
	SetPassword(password string) types.Error
	Profile() (types.UserProfile, types.Error)
	SetDisplayName(name string, by User) types.Error
	SetAvatarUrl(url string, by User) types.Error
}

type TokenService interface {
	NewAccessToken(types.UserId) (Token, types.Error)
	ParseAccessToken(token string) (Token, types.Error)
}

type Token interface {
	fmt.Stringer
	UserId() types.UserId
}

type UserStore interface {
	CreateUser(types.UserId) types.Error
	UserExists(types.UserId) types.Error
	SetUserPasswordHash(id types.UserId, hash string) types.Error
	UserPasswordHash(types.UserId) (string, types.Error)
	SetUserDisplayName(id types.UserId, displayName string) types.Error
	SetUserAvatarUrl(id types.UserId, avatarUrl string) types.Error
	UserProfile(types.UserId) (types.UserProfile, types.Error)
}

type RoomStore interface {
	CreateRoom(domain string) (types.RoomId, types.Error)
	RoomExists(types.RoomId) (bool, types.Error)
	AddRoomMessage(types.RoomId, types.UserId, types.TypedContent) (*types.Message, types.Error)
	SetRoomState(roomId types.RoomId, userId types.UserId, content types.TypedContent, stateKey string) (*types.State, types.Error)
	RoomState(roomId types.RoomId, eventType, stateKey string) (*types.State, types.Error)
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
}
