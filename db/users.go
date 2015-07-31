package db

import (
	"errors"
	"time"
)

type Presence string

const (
	PresenceOnline      Presence = "online"
	PresenceOffline              = "offline"
	PresenceAvailable            = "free_for_chat"
	PresenceUnavailable          = "unavailable"
)

type User struct {
	id            string
	passwordHash  string
	displayName   string
	avatarUrl     string
	presence      Presence
	statusMessage string
	lastActive    time.Time
}

var userTable = map[string]*User{}

func CreateUser(id string) error {
	if userTable[id] != nil {
		return errors.New("user already exists")
	}
	userTable[id] = &User{
		id:         id,
		presence:   "offline",
		lastActive: time.Now(),
	}
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
	user.passwordHash = hash
	return nil
}

func GetUserPasswordHash(id string) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", errors.New("user not found")
	}
	return user.passwordHash, nil
}

func SetUserDisplayName(id string, displayName string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.displayName = displayName
	return nil
}

func GetUserDisplayName(id string) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", errors.New("user not found")
	}
	return user.displayName, nil
}

func SetUserAvatarUrl(id string, avatarUrl string) error {
	user := userTable[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.avatarUrl = avatarUrl
	return nil
}

func GetUserAvatarUrl(id string) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", errors.New("user not found")
	}
	return user.avatarUrl, nil
}
