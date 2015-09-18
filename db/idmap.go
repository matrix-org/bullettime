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
	"time"

	"github.com/matrix-org/bullettime/interfaces"
	"github.com/matrix-org/bullettime/types"
)

type idMapDb struct { // always lock in the same order as below
	reverseMappingLock sync.RWMutex
	reverseMapping     map[types.Id][]types.Id

	mappingLock sync.RWMutex
	mapping     map[types.Id]types.Id
	reserved    map[types.Id]struct{}
}

func NewIdMapDb() (interfaces.IdMapStore, error) {
	return &idMapDb{
		reverseMapping: map[types.Id][]types.Id{},
		mapping:        map[types.Id]types.Id{},
		reserved:       map[types.Id]struct{}{},
	}, nil
}

func (db *idMapDb) Reserve(from types.Id) types.Error {
	db.mappingLock.Lock()
	defer db.mappingLock.Unlock()
	if _, ok := db.mapping[from]; ok {
		return types.RoomInUseError("id from '" + from.String() + "' already exists")
	}
	if _, ok := db.reserved[from]; ok {
		return types.RoomInUseError("id from '" + from.String() + "' already reserved")
	}
	db.reserved[from] = struct{}{}
	go func() {
		time.Sleep(time.Second * 10)
		delete(db.reserved, from)
	}()
	return nil
}

func (db *idMapDb) Claim(from types.Id, to types.Id) types.Error {
	db.mappingLock.Lock()
	defer db.mappingLock.Unlock()
	if _, ok := db.mapping[from]; ok {
		return types.RoomInUseError("id from '" + from.String() + "' already exists")
	}
	if _, ok := db.reserved[from]; !ok {
		return types.RoomInUseError("id from '" + from.String() + "' was not reserved")
	}
	delete(db.reserved, from)
	db.mapping[from] = to

	db.reverseMappingLock.Lock()
	defer db.reverseMappingLock.Unlock()
	db.reverseMapping[to] = append(db.reverseMapping[to], from)
	return nil
}

func (db *idMapDb) Put(from types.Id, to types.Id) types.Error {
	db.mappingLock.Lock()
	defer db.mappingLock.Unlock()
	if _, ok := db.mapping[from]; ok {
		return types.RoomInUseError("id from '" + from.String() + "' already exists")
	}
	if _, ok := db.reserved[from]; ok {
		return types.RoomInUseError("id from '" + from.String() + "' is reserved")
	}
	db.mapping[from] = to

	db.reverseMappingLock.Lock()
	defer db.reverseMappingLock.Unlock()
	db.reverseMapping[to] = append(db.reverseMapping[to], from)

	return nil
}

func (db *idMapDb) Delete(from types.Id, to types.Id) types.Error {
	db.mappingLock.Lock()
	defer db.mappingLock.Unlock()
	if _, ok := db.mapping[from]; !ok {
		return types.NotFoundError("id from '" + from.String() + "' doesn't exist")
	}
	delete(db.mapping, from)

	db.reverseMappingLock.Lock()
	defer db.reverseMappingLock.Unlock()

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
	return nil
}

func (db *idMapDb) Lookup(from types.Id) (*types.Id, types.Error) {
	db.mappingLock.RLock()
	defer db.mappingLock.RUnlock()
	if to, ok := db.mapping[from]; ok {
		return &to, nil
	}
	return nil, nil
}

func (db *idMapDb) ReverseLookup(to types.Id) ([]types.Id, types.Error) {
	db.reverseMappingLock.RLock()
	defer db.reverseMappingLock.RUnlock()
	return db.reverseMapping[to], nil
}
