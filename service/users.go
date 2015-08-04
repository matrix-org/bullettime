package service

import (
	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserService() (interfaces.UserService, error) {
	return userService{}, nil
}

type userService struct{}

type userInfo struct {
	id types.UserId
}

func (u userService) GetUser(id types.UserId) (interfaces.User, error) {
	if err := db.UserExists(id); err != nil {
		return nil, err
	}
	return userInfo{id: id}, nil
}

func (u userService) CreateUser(id types.UserId) (interfaces.User, error) {
	if err := db.CreateUser(id); err != nil {
		return nil, err
	}
	return userInfo{id: id}, nil
}

func (u userInfo) Id() types.UserId {
	return u.id
}

func (u userInfo) VerifyPassword(password string) error {
	hash, err := db.GetUserPasswordHash(u.id)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (u userInfo) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	if err := db.SetUserPasswordHash(u.id, string(hash)); err != nil {
		return err
	}
	return nil
}

func (u userInfo) GetProfile() (types.UserProfile, error) {
	return db.GetUserProfile(u.id)
}

func (u userInfo) SetDisplayName(displayName string, doneBy interfaces.User) error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserDisplayName(u.id, displayName)
}

func (u userInfo) SetAvatarUrl(avatarUrl string, doneBy interfaces.User) error {
	if u != doneBy {
		return types.ForbiddenError("can't change the display name of other users")
	}
	return db.SetUserAvatarUrl(u.id, avatarUrl)
}
