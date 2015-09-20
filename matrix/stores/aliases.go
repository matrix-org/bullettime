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
	"unsafe"

	ci "github.com/matrix-org/bullettime/core/interfaces"
	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	"github.com/matrix-org/bullettime/matrix/types"
)

type aliasStore struct {
	idMap ci.IdMap
}

func NewAliasStore(idMap ci.IdMap) (interfaces.AliasStore, error) {
	return &aliasStore{idMap}, nil
}

func (s *aliasStore) AddAlias(alias ct.Alias, room ct.RoomId) types.Error {
	inserted, err := s.idMap.Insert(ct.Id(alias), ct.Id(room))
	if err != nil {
		return types.InternalError(err)
	}
	if !inserted {
		return types.RoomInUseError("room alias '" + alias.String() + "' already exists")
	}
	return nil
}

func (s *aliasStore) RemoveAlias(alias ct.Alias, room ct.RoomId) types.Error {
	deleted, err := s.idMap.Delete(ct.Id(alias), ct.Id(room))
	if err != nil {
		return types.InternalError(err)
	}
	if !deleted {
		return types.NotFoundError("room alias '" + alias.String() + "' doesn't exist")
	}
	return nil
}

func (s *aliasStore) Room(alias ct.Alias) (*ct.RoomId, types.Error) {
	room, err := s.idMap.Lookup(ct.Id(alias))
	if err != nil {
		return nil, types.InternalError(err)
	}
	return (*ct.RoomId)(room), nil
}

func (s *aliasStore) Aliases(room ct.RoomId) ([]ct.Alias, types.Error) {
	ids, err := s.idMap.ReverseLookup(ct.Id(room))
	if err != nil {
		return nil, types.InternalError(err)
	}
	aliases := *(*[]ct.Alias)(unsafe.Pointer(&ids))
	return aliases, nil
}
