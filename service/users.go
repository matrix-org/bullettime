package service

import (
	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserService() (interfaces.UserService, types.Error) {
	return userService{}, nil
}

type userService struct{}

type userInfo struct {
	id types.UserId
}

func (u userService) User(id types.UserId) (interfaces.User, types.Error) {
	if err := db.UserExists(id); err != nil {
		return nil, err
	}
	return userInfo{id: id}, nil
}

func (u userService) CreateUser(id types.UserId) (interfaces.User, types.Error) {
	if err := db.CreateUser(id); err != nil {
		return nil, err
	}
	return userInfo{id: id}, nil
}

func (u userInfo) Id() types.UserId {
	return u.id
}

func (u userInfo) VerifyPassword(password string) types.Error {
	hash, err := db.UserPasswordHash(u.id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return types.ForbiddenError("invalid credentials")
	}
	return nil
}

func (u userInfo) SetPassword(password string) types.Error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return types.ServerError("failed to generate password: " + err.Error())
	}
	if err := db.SetUserPasswordHash(u.id, string(hash)); err != nil {
		return err
	}
	return nil
}

func (u userInfo) Profile() (types.UserProfile, types.Error) {
	return db.UserProfile(u.id)
}

func (u userInfo) SetDisplayName(displayName string, doneBy interfaces.User) types.Error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserDisplayName(u.id, displayName)
}

func (u userInfo) SetAvatarUrl(avatarUrl string, doneBy interfaces.User) types.Error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserAvatarUrl(u.id, avatarUrl)
}
