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

	"github.com/matrix-org/bullettime/core/interfaces"
	"github.com/matrix-org/bullettime/core/types"
)

type idMapDb struct { // always lock in the same order as below
	sync.RWMutex
	mapping        map[types.Id]types.Id
	reverseMapping map[types.Id][]types.Id
}

func NewIdMap() (interfaces.IdMap, error) {
	return &idMapDb{
		mapping:        map[types.Id]types.Id{},
		reverseMapping: map[types.Id][]types.Id{},
	}, nil
}

func (db *idMapDb) Insert(key types.Id, value types.Id) (inserted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[key]; ok {
		return false, nil
	}
	db.mapping[key] = value
	db.reverseMapping[value] = append(db.reverseMapping[value], key)
	return true, nil
}

func (db *idMapDb) Replace(key types.Id, value types.Id) (replaced bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[key]; !ok {
		return false, nil
	}
	db.mapping[key] = value
	return true, nil
}

func (db *idMapDb) Put(key types.Id, value types.Id) types.Error {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[key]; ok {
		db.reverseMapping[value] = append(db.reverseMapping[value], key)
	}
	db.mapping[key] = value
	return nil
}

func (db *idMapDb) Delete(key types.Id, value types.Id) (deleted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.mapping[key]; !ok {
		return false, nil
	}
	delete(db.mapping, key)

	reverseMapping := db.reverseMapping[value]
	l := len(reverseMapping)
	for i := 0; i < l; i += 1 {
		if reverseMapping[i] == key {
			reverseMapping[i] = reverseMapping[l-1]
			reverseMapping[l-1] = types.Id{}
			reverseMapping = reverseMapping[:l-1]
			break
		}
	}
	db.reverseMapping[value] = reverseMapping
	return true, nil
}

func (db *idMapDb) Lookup(key types.Id) (*types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	if value, ok := db.mapping[key]; ok {
		return &value, nil
	}
	return nil, nil
}

func (db *idMapDb) ReverseLookup(value types.Id) ([]types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.reverseMapping[value], nil
}
