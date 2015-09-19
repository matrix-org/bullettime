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

	"github.com/matrix-org/bullettime/core/types"
	matrixTypes "github.com/matrix-org/bullettime/matrix/types"
)

type stateStore struct { // always lock in the same order as below
	sync.RWMutex
	buckets map[types.Id]*bucket
}

type bucket struct { // always lock in the same order as below
	sync.RWMutex
	states map[string][]byte
}

type state struct {
	key   string
	value []byte
}

func (s state) Key() string {
	return s.key
}
func (s state) Value() []byte {
	return s.value
}

func NewStateStore() (StateStore, error) {
	return &stateStore{
		buckets: map[types.Id]*bucket{},
	}, nil
}

func (db *stateStore) CreateBucket(id types.Id) (bool, types.Error) {
	db.Lock()
	defer db.Unlock()
	if db.buckets[id] != nil {
		return true, nil
	}
	db.buckets[id] = &bucket{
		states: map[string][]byte{},
	}
	return false, nil
}

func (db *stateStore) BucketExists(id types.Id) (bool, types.Error) {
	db.RLock()
	defer db.RUnlock()
	if db.buckets[id] == nil {
		return false, nil
	}
	return true, nil
}

func (db *stateStore) SetState(id types.Id, key string, value []byte) ([]byte, types.Error) {
	db.RLock()
	defer db.RUnlock()
	bucket := db.buckets[id]
	if bucket == nil {
		return nil, matrixTypes.NotFoundError("bucket '" + id.String() + "' doesn't exist")
	}
	bucket.Lock()
	defer bucket.Unlock()
	oldValue := bucket.states[key]
	if len(value) == 0 {
		delete(bucket.states, key)
	} else {
		bucket.states[key] = value
	}

	return oldValue, nil
}

func (db *stateStore) State(id types.Id, key string) ([]byte, types.Error) {
	db.RLock()
	defer db.RUnlock()
	bucket := db.buckets[id]
	if bucket == nil {
		return nil, matrixTypes.NotFoundError("bucket '" + id.String() + "' doesn't exist")
	}
	bucket.RLock()
	defer bucket.RUnlock()
	value := bucket.states[key]
	return value, nil
}

func (db *stateStore) States(id types.Id) ([]State, types.Error) {
	db.RLock()
	defer db.RUnlock()
	bucket := db.buckets[id]
	if bucket == nil {
		return nil, matrixTypes.NotFoundError("bucket '" + id.String() + "' doesn't exist")
	}
	bucket.RLock()
	defer bucket.RUnlock()
	states := make([]State, 0, len(bucket.states))
	for key, value := range bucket.states {
		states = append(states, state{key, value})
	}
	return states, nil
}
