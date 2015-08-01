package db

import (
	"errors"
	"time"

	"github.com/Rugvip/bullettime/types"
)

type User struct {
	Id types.UserId
	types.UserProfile
	types.UserPresence
	PasswordHash string `json:"-"`
}

var userTable = map[types.UserId]*User{}

func GetUser(id types.UserId) *User {
	return userTable[id]
}

func CreateUser(id types.UserId) error {
	if userTable[id] != nil {
		return errors.New("user already exists")
	}
	user := new(User)
	user.Id = id
	user.LastActive = types.LastActive(time.Now())
	userTable[id] = user
	return nil
}

func UserExists(id types.UserId) error {
	if userTable[id] == nil {
		return errors.New("user not found")
	}
	return nil
}

func SetUserPasswordHash(id types.UserId, hash string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.PasswordHash = hash
	return nil
}

func GetUserPasswordHash(id types.UserId) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", errors.New("user not found")
	}
	return user.PasswordHash, nil
}

func SetUserDisplayName(id types.UserId, displayName string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.DisplayName = displayName
	return nil
}

func SetUserAvatarUrl(id types.UserId, avatarUrl string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.AvatarUrl = avatarUrl
	return nil
}

func GetUserProfile(id types.UserId) (types.UserProfile, error) {
	user := userTable[id]
	if user == nil {
		return types.UserProfile{}, errors.New("user not found")
	}
	return user.UserProfile, nil
}
