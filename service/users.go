// Copyright 2015  Ericsson AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func (s userService) UserExists(user, caller types.UserId) (bool, types.Error) {
	return s.users.UserExists(user)
}

func (s userService) CreateUser(id types.UserId) types.Error {
	return s.users.CreateUser(id)
}

func (s userService) VerifyPassword(user types.UserId, password string) (bool, types.Error) {
	hash, err := s.users.UserPasswordHash(user)
	if err != nil {
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false, nil
	}
	return true, nil
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
