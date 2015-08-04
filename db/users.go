package db

import (
	"github.com/Rugvip/bullettime/interfaces"

	"github.com/Rugvip/bullettime/types"
)

type userDb struct {
	users map[types.UserId]*dbUser
}

func NewUserDb() (interfaces.UserStore, types.Error) {
	return userDb{
		users: map[types.UserId]*dbUser{},
	}, nil
}

type dbUser struct {
	types.UserId
	types.UserProfile
	PasswordHash string `json:"-"`
}

func (db userDb) CreateUser(id types.UserId) types.Error {
	if db.users[id] != nil {
		return types.UserInUseError("user '" + id.String() + "' already exists")
	}
	user := new(dbUser)
	user.UserId = id
	db.users[id] = user
	return nil
}

func (db userDb) UserExists(id types.UserId) types.Error {
	if db.users[id] == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return nil
}

func (db userDb) SetUserPasswordHash(id types.UserId, hash string) types.Error {
	user := db.users[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.PasswordHash = hash
	return nil
}

func (db userDb) UserPasswordHash(id types.UserId) (string, types.Error) {
	user := db.users[id]
	if user == nil {
		return "", types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.PasswordHash, nil
}

func (db userDb) SetUserDisplayName(id types.UserId, displayName string) types.Error {
	user := db.users[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.DisplayName = displayName
	return nil
}

func (db userDb) SetUserAvatarUrl(id types.UserId, avatarUrl string) types.Error {
	user := db.users[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.AvatarUrl = avatarUrl
	return nil
}

func (db userDb) UserProfile(id types.UserId) (types.UserProfile, types.Error) {
	user := db.users[id]
	if user == nil {
		return types.UserProfile{}, types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.UserProfile, nil
}
