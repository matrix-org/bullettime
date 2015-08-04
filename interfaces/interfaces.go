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
