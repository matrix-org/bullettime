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

type idMultiMap struct { // always lock in the same order as below
	sync.RWMutex
	mapping        map[types.Id][]types.Id
	reverseMapping map[types.Id][]types.Id
	entries        map[entryKey]struct{}
}

type entryKey struct {
	key   types.Id
	value types.Id
}

func NewIdMultiMapStore() (interfaces.IdMultiMapStore, error) {
	return &idMultiMap{
		mapping:        map[types.Id][]types.Id{},
		reverseMapping: map[types.Id][]types.Id{},
		entries:        map[entryKey]struct{}{},
	}, nil
}

func (db *idMultiMap) Put(key types.Id, value types.Id) (inserted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	entry := entryKey{key, value}
	if _, ok := db.entries[entry]; ok {
		return false, nil
	}
	db.entries[entry] = struct{}{}
	db.mapping[key] = append(db.mapping[key], value)
	db.reverseMapping[value] = append(db.reverseMapping[value], key)
	return true, nil
}

func (db *idMultiMap) Delete(key types.Id, value types.Id) (deleted bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	entry := entryKey{key, value}
	if _, ok := db.entries[entry]; !ok {
		return false, nil
	}
	mapping := db.mapping[key]
	for i, l := 0, len(mapping); i < l; i += 1 {
		if mapping[i] == value {
			mapping[i] = mapping[l-1]
			mapping[l-1] = types.Id{}
			mapping = mapping[:l-1]
			break
		}
	}
	reverseMapping := db.reverseMapping[value]
	for i, l := 0, len(reverseMapping); i < l; i += 1 {
		if reverseMapping[i] == key {
			reverseMapping[i] = reverseMapping[l-1]
			reverseMapping[l-1] = types.Id{}
			reverseMapping = reverseMapping[:l-1]
			db.reverseMapping[value] = reverseMapping
			break
		}
	}
	delete(db.entries, entry)
	return true, nil
}

func (db *idMultiMap) Contains(key types.Id, value types.Id) (exists bool, err types.Error) {
	db.Lock()
	defer db.Unlock()
	entry := entryKey{key, value}
	if _, ok := db.entries[entry]; ok {
		return true, nil
	}
	return false, nil
}

func (db *idMultiMap) Lookup(key types.Id) ([]types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.mapping[key], nil
}

func (db *idMultiMap) ReverseLookup(value types.Id) ([]types.Id, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.reverseMapping[value], nil
}
