package db

import (
	"errors"
	"time"

	"github.com/Rugvip/bullettime/types"
)

type User struct {
	types.UserInfo
	types.UserPresence
	PasswordHash string `json:"-"`
}

var userTable = map[string]*User{}

func GetUser(id string) *User {
	return userTable[id]
}

func CreateUser(id string) error {
	if userTable[id] != nil {
		return errors.New("user already exists")
	}
	user := new(User)
	user.Id = id
	user.LastActive = types.LastActive(time.Now())
	userTable[id] = user
	return nil
}

func UserExists(id string) error {
	if userTable[id] == nil {
		return errors.New("user not found")
	}
	return nil
}

func SetUserPasswordHash(id, hash string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.PasswordHash = hash
	return nil
}

func GetUserPasswordHash(id string) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", errors.New("user not found")
	}
	return user.PasswordHash, nil
}

func SetUserDisplayName(id string, displayName string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.DisplayName = displayName
	return nil
}

func SetUserAvatarUrl(id string, avatarUrl string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.AvatarUrl = avatarUrl
	return nil
}

func GetUserInfo(id string) (types.UserInfo, error) {
	user := userTable[id]
	if user == nil {
		return types.UserInfo{}, errors.New("user not found")
	}
	return user.UserInfo, nil
}
