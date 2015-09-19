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
	"github.com/matrix-org/bullettime/utils"
)

type roomDb struct { // always lock in the same order as below
	roomsLock sync.RWMutex
	rooms     map[types.RoomId]*dbRoom
}

func NewRoomDb() (interfaces.RoomStore, error) {
	return &roomDb{
		rooms: map[types.RoomId]*dbRoom{},
	}, nil
}

type stateId struct {
	EventType string
	StateKey  string
}

type dbRoom struct { // always lock in the same order as below
	id        types.RoomId
	stateLock sync.RWMutex
	states    map[stateId]*types.State
}

func (db *roomDb) CreateRoom(domain string) (types.RoomId, types.Error) {
	db.roomsLock.Lock()
	defer db.roomsLock.Unlock()
	id := types.NewRoomId("", domain)
	for {
		id.Id = utils.RandomString(16)
		if db.rooms[id] == nil {
			break
		}
	}
	db.rooms[id] = &dbRoom{
		id:     id,
		states: map[stateId]*types.State{},
	}
	return id, nil
}

func (db *roomDb) RoomExists(id types.RoomId) (bool, types.Error) {
	db.roomsLock.RLock()
	defer db.roomsLock.RUnlock()
	if db.rooms[id] == nil {
		return false, nil
	}
	return true, nil
}

func (db *roomDb) SetRoomState(roomId types.RoomId, userId types.UserId, content types.TypedContent, stateKey string) (*types.State, types.Error) {
	db.roomsLock.RLock()
	defer db.roomsLock.RUnlock()
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	var eventId = types.DeriveEventId(utils.RandomString(16), types.Id(userId))
	stateId := stateId{content.GetEventType(), stateKey}

	state := new(types.State)
	state.EventId = eventId
	state.RoomId = roomId
	state.UserId = userId
	state.EventType = content.GetEventType()
	state.StateKey = stateKey
	state.Timestamp = types.Timestamp{time.Now()}
	state.Content = content
	state.OldState = (*types.OldState)(room.states[stateId])

	room.stateLock.Lock()
	defer room.stateLock.Unlock()
	room.states[stateId] = state

	return state, nil
}

func (db *roomDb) RoomState(roomId types.RoomId, eventType, stateKey string) (*types.State, types.Error) {
	db.roomsLock.RLock()
	defer db.roomsLock.RUnlock()
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	room.stateLock.RLock()
	defer room.stateLock.RUnlock()
	state := room.states[stateId{eventType, stateKey}]
	return state, nil
}

func (db *roomDb) EntireRoomState(roomId types.RoomId) ([]*types.State, types.Error) {
	db.roomsLock.RLock()
	defer db.roomsLock.RUnlock()
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	room.stateLock.RLock()
	defer room.stateLock.RUnlock()
	states := make([]*types.State, 0, len(room.states))
	for _, state := range room.states {
		states = append(states, state)
	}
	return states, nil
}
