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
	"github.com/matrix-org/bullettime/interfaces"

	"github.com/matrix-org/bullettime/types"
)

type userDb struct {
	StateStore
}

const passwordHashKey = "pw_hash"

func NewUserDb(stateStore StateStore) (interfaces.UserStore, error) {
	return &userDb{stateStore}, nil
}

func (db *userDb) CreateUser(id types.UserId) (bool, types.Error) {
	return db.CreateBucket(types.Id(id))
}

func (db *userDb) UserExists(id types.UserId) (bool, types.Error) {
	return db.BucketExists(types.Id(id))
}

func (db *userDb) SetUserPasswordHash(id types.UserId, hash string) types.Error {
	_, err := db.SetState(types.Id(id), passwordHashKey, []byte(hash))
	return err
}

func (db *userDb) UserPasswordHash(id types.UserId) (string, types.Error) {
	value, err := db.State(types.Id(id), passwordHashKey)
	if err != nil {
		return "", err
	}
	return string(value), nil
}
