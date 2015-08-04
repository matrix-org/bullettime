package service

import (
	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/types"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id types.UserId
}

func GetUser(id types.UserId) (User, error) {
	if err := db.UserExists(id); err != nil {
		return User{}, err
	}
	return User{id: id}, nil
}

func CreateUser(id types.UserId) (User, error) {
	if err := db.CreateUser(id); err != nil {
		return User{}, err
	}
	return User{id: id}, nil
}

func (u User) Id() types.UserId {
	return u.id
}

func (u User) VerifyPassword(password string) error {
	hash, err := db.GetUserPasswordHash(u.id)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (u User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	if err := db.SetUserPasswordHash(u.id, string(hash)); err != nil {
		return err
	}
	return nil
}

func (u User) GetProfile() (types.UserProfile, error) {
	return db.GetUserProfile(u.id)
}

func (u User) SetDisplayName(displayName string, doneBy User) error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserDisplayName(u.id, displayName)
}

func (u User) SetAvatarUrl(avatarUrl string, doneBy User) error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserAvatarUrl(u.id, avatarUrl)
}
