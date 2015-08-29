package service

import (
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserService(
	users interfaces.UserStore,
) (interfaces.UserService, error) {
	return userService{
		users,
	}, nil
}

type userService struct {
	users interfaces.UserStore
}

func (s userService) UserExists(user, caller types.UserId) types.Error {
	exists, err := s.users.UserExists(user)
	if err != nil {
		return err
	}
	if !exists {
		return types.NotFoundError("user '" + user.String() + "' doesn't exist")
	}
	return nil
}

func (s userService) CreateUser(id types.UserId) types.Error {
	return s.users.CreateUser(id)
}

func (s userService) VerifyPassword(user types.UserId, password string) types.Error {
	hash, err := s.users.UserPasswordHash(user)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return types.ForbiddenError("invalid credentials")
	}
	return nil
}

func (s userService) SetPassword(user, caller types.UserId, password string) types.Error {
	if user != caller {
		return types.ForbiddenError("can't change the password of other users")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return types.ServerError("failed to generate password: " + err.Error())
	}
	if err := s.users.SetUserPasswordHash(user, string(hash)); err != nil {
		return err
	}
	return nil
}
