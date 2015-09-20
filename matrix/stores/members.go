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
	"fmt"
	"unsafe"

	ci "github.com/matrix-org/bullettime/core/interfaces"
	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	"github.com/matrix-org/bullettime/matrix/types"
)

type memberStore struct {
	idMap ci.IdMultiMapStore
}

func NewMembershipStore(idMultiMapStore ci.IdMultiMapStore) (interfaces.MembershipStore, error) {
	return &memberStore{idMultiMapStore}, nil
}

func (db *memberStore) AddMember(roomId ct.RoomId, userId ct.UserId) types.Error {
	inserted, err := db.idMap.Put(ct.Id(roomId), ct.Id(userId))
	if err != nil {
		return types.InternalError(err)
	}
	if !inserted {
		msg := fmt.Sprintf("user %s is already a member of the room %s", userId, roomId)
		return types.ServerError(msg)
	}
	return nil
}

func (db *memberStore) RemoveMember(roomId ct.RoomId, userId ct.UserId) types.Error {
	deleted, err := db.idMap.Delete(ct.Id(roomId), ct.Id(userId))
	if err != nil {
		return types.InternalError(err)
	}
	if !deleted {
		msg := fmt.Sprintf("user %s is not a member of the room %s", userId, roomId)
		return types.ServerError(msg)
	}
	return nil
}

func (db *memberStore) Rooms(userId ct.UserId) ([]ct.RoomId, types.Error) {
	ids, err := db.idMap.ReverseLookup(ct.Id(userId))
	rooms := *(*[]ct.RoomId)(unsafe.Pointer(&ids))
	return rooms, types.InternalError(err)
}

func (db *memberStore) Users(roomId ct.RoomId) ([]ct.UserId, types.Error) {
	ids, err := db.idMap.Lookup(ct.Id(roomId))
	users := *(*[]ct.UserId)(unsafe.Pointer(&ids))
	return users, types.InternalError(err)
}

func (db *memberStore) Peers(userId ct.UserId) (map[ct.UserId]struct{}, types.Error) {
	idSet, err := db.idMap.ReverseLinkUnionLookup(ct.Id(userId))
	peers := *(*map[ct.UserId]struct{})(unsafe.Pointer(&idSet))
	return peers, types.InternalError(err)
}
