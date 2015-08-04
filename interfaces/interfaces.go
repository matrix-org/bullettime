package interfaces

import (
	"github.com/Rugvip/bullettime/types"
)

type RoomService interface {
	GetRoom(types.RoomId) (Room, error)
	CreateRoom(hostname string, creator User, desc *types.RoomDescription) (Room, *types.Alias, error)
}

type Room interface {
	Id() types.RoomId
	AddMessage(user User, content types.TypedContent) (*types.Event, error)
	SetState(user User, content types.TypedContent, stateKey string) (*types.State, error)
	GetState(user User, eventType, stateKey string) (*types.State, error)
}

type UserService interface {
	GetUser(types.UserId) (User, error)
	CreateUser(types.UserId) (User, error)
}

type User interface {
	Id() types.UserId
	VerifyPassword(password string) error
	SetPassword(password string) error
	GetProfile() (types.UserProfile, error)
	SetDisplayName(name string, by User) error
	SetAvatarUrl(url string, by User) error
}
