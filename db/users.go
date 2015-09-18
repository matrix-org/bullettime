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

package db

import (
	"sync"

	"github.com/matrix-org/bullettime/interfaces"

	"github.com/matrix-org/bullettime/types"
)

type userDb struct {
	sync.RWMutex
	users map[types.UserId]*dbUser
}

func NewUserDb() (interfaces.UserStore, error) {
	return &userDb{
		users: map[types.UserId]*dbUser{},
	}, nil
}

type dbUser struct {
	sync.RWMutex
	types.UserId
	types.UserProfile
	PasswordHash string `json:"-"`
}

func (db *userDb) CreateUser(id types.UserId) types.Error {
	db.Lock()
	defer db.Unlock()
	if db.users[id] != nil {
		return types.UserInUseError("user '" + id.String() + "' already exists")
	}
	user := new(dbUser)
	user.UserId = id
	db.users[id] = user
	return nil
}

func (db *userDb) UserExists(id types.UserId) (bool, types.Error) {
	db.RLock()
	defer db.RUnlock()
	if db.users[id] == nil {
		return false, nil
	}
	return true, nil
}

func (db *userDb) SetUserPasswordHash(id types.UserId, hash string) types.Error {
	db.RLock()
	defer db.RUnlock()
	user := db.users[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.Lock()
	defer user.Unlock()
	user.PasswordHash = hash
	return nil
}

func (db *userDb) UserPasswordHash(id types.UserId) (string, types.Error) {
	db.RLock()
	defer db.RUnlock()
	user := db.users[id]
	if user == nil {
		return "", types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.RLock()
	defer user.RUnlock()
	return user.PasswordHash, nil
}

// func (db userDb) SetUserDisplayName(id types.UserId, displayName string) types.Error {
// 	db.RLock()
// 	defer db.RUnlock()
// 	user := db.users[id]
// 	if user == nil {
// 		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
// 	}
// 	user.Lock()
// 	defer user.Unlock()
// 	user.DisplayName = displayName
// 	return nil
// }

// func (db userDb) SetUserAvatarUrl(id types.UserId, avatarUrl string) types.Error {
// 	db.RLock()
// 	defer db.RUnlock()
// 	user := db.users[id]
// 	if user == nil {
// 		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
// 	}
// 	user.Lock()
// 	defer user.Unlock()
// 	user.AvatarUrl = avatarUrl
// 	return nil
// }

// func (db userDb) UserProfile(id types.UserId) (types.UserProfile, types.Error) {
// 	db.RLock()
// 	defer db.RUnlock()
// 	user := db.users[id]
// 	if user == nil {
// 		return types.UserProfile{}, types.NotFoundError("user '" + id.String() + "' doesn't exist")
// 	}
// 	user.RLock()
// 	defer user.RUnlock()
// 	return user.UserProfile, nil
// }
