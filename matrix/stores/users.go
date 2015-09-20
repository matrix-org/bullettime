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

package stores

import (
	ci "github.com/matrix-org/bullettime/core/interfaces"
	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
)

type userDb struct {
	ci.StateStore
}

const passwordHashKey = "pw_hash"

func NewUserDb(stateStore ci.StateStore) (interfaces.UserStore, error) {
	return &userDb{stateStore}, nil
}

func (db *userDb) CreateUser(id ct.UserId) (bool, ct.Error) {
	return db.CreateBucket(ct.Id(id))
}

func (db *userDb) UserExists(id ct.UserId) (bool, ct.Error) {
	return db.BucketExists(ct.Id(id))
}

func (db *userDb) SetUserPasswordHash(id ct.UserId, hash string) ct.Error {
	_, err := db.SetState(ct.Id(id), passwordHashKey, []byte(hash))
	return err
}

func (db *userDb) UserPasswordHash(id ct.UserId) (string, ct.Error) {
	value, err := db.State(ct.Id(id), passwordHashKey)
	if err != nil {
		return "", err
	}
	return string(value), nil
}
