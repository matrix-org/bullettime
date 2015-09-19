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

type idMapDb struct { // always lock in the same order as below
	sync.RWMutex
	mapping        map[types.Id]types.Id
	reverseMapping map[types.Id][]types.Id
}

func NewIdMapDb() (interfaces.IdMapStore, error) {
	return &idMapDb{
		mapping:        map[types.Id]types.Id{},
		reverseMapping: map[types.Id][]types.Id{},
	}, nil
}

func (db *idMapDb) Insert(from types.Id, to types.Id) (inserted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[from]; ok {
		return false, nil
	}
	db.mapping[from] = to
	db.reverseMapping[to] = append(db.reverseMapping[to], from)
	return true, nil
}

func (db *idMapDb) Replace(from types.Id, to types.Id) (replaced bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[from]; !ok {
		return false, nil
	}
	db.mapping[from] = to
	return true, nil
}

func (db *idMapDb) Put(from types.Id, to types.Id) types.Error {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[from]; ok {
		db.reverseMapping[to] = append(db.reverseMapping[to], from)
	}
	db.mapping[from] = to
	return nil
}

func (db *idMapDb) Delete(from types.Id, to types.Id) (deleted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[from]; !ok {
		return false, nil
	}
	delete(db.mapping, from)

	reverseMapping := db.reverseMapping[to]
	l := len(reverseMapping)
	for i := 0; i < l; i += 1 {
		if reverseMapping[i] == from {
			reverseMapping[i] = reverseMapping[l-1]
			reverseMapping[l-1] = types.Id{}
			reverseMapping = reverseMapping[:l-1]
			break
		}
	}
	db.reverseMapping[to] = reverseMapping
	return true, nil
}

func (db *idMapDb) Lookup(from types.Id) (*types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	if to, ok := db.mapping[from]; ok {
		return &to, nil
	}
	return nil, nil
}

func (db *idMapDb) ReverseLookup(to types.Id) ([]types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.reverseMapping[to], nil
}
