package service

import (
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserService(userStore interfaces.UserStore) (interfaces.UserService, types.Error) {
	return userService{userStore}, nil
}

type userService struct {
	db interfaces.UserStore
}

type userInfo struct {
	id      types.UserId
	service userService
}

func (u userService) User(id types.UserId) (interfaces.User, types.Error) {
	if err := u.db.UserExists(id); err != nil {
		return nil, err
	}
	return userInfo{id, u}, nil
}

func (u userService) CreateUser(id types.UserId) (interfaces.User, types.Error) {
	if err := u.db.CreateUser(id); err != nil {
		return nil, err
	}
	return userInfo{id, u}, nil
}

func (u userInfo) Id() types.UserId {
	return u.id
}

func (u userInfo) VerifyPassword(password string) types.Error {
	hash, err := u.service.db.UserPasswordHash(u.id)
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
	if err := u.service.db.SetUserPasswordHash(u.id, string(hash)); err != nil {
		return err
	}
	return nil
}

func (u userInfo) Profile() (types.UserProfile, types.Error) {
	return u.service.db.UserProfile(u.id)
}

func (u userInfo) SetDisplayName(displayName string, doneBy interfaces.User) types.Error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return u.service.db.SetUserDisplayName(u.id, displayName)
}

func (u userInfo) SetAvatarUrl(avatarUrl string, doneBy interfaces.User) types.Error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return u.service.db.SetUserAvatarUrl(u.id, avatarUrl)
}
