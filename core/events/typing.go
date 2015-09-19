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

package events

import (
	"sync"
	"sync/atomic"

	"github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	matrixTypes "github.com/matrix-org/bullettime/matrix/types"
)

type typingStream struct {
	lock           sync.RWMutex
	states         map[types.RoomId]*indexedTypingState
	max            uint64
	members        interfaces.MembershipStore
	asyncEventSink interfaces.AsyncEventSink
}

type indexedTypingState struct {
	event matrixTypes.TypingEvent
	index uint64
}

func (m *indexedTypingState) Event() types.Event {
	return &m.event
}

func (s *indexedTypingState) Index() uint64 {
	return s.index
}

func NewTypingStream(
	members interfaces.MembershipStore,
	asyncEventSink interfaces.AsyncEventSink,
) (interfaces.TypingStream, error) {
	return &typingStream{
		states:         map[types.RoomId]*indexedTypingState{},
		members:        members,
		asyncEventSink: asyncEventSink,
	}, nil
}

func (s *typingStream) SetTyping(room types.RoomId, user types.UserId, typing bool) types.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	state := s.states[room]
	index := atomic.AddUint64(&s.max, 1) - 1
	if state == nil {
		state = &indexedTypingState{index: index}
		state.event.RoomId = room
		state.event.EventType = matrixTypes.EventTypeTyping
		s.states[room] = state
	} else {
		state.index = index
	}
	userIds := state.event.Content.UserIds
	if typing {
		for _, member := range userIds {
			if member == user {
				return nil
			}
		}
		state.event.Content.UserIds = append(userIds, user)
	} else {
		for i, member := range userIds {
			if member == user {
				userIds[i] = userIds[len(userIds)-1]
				state.event.Content.UserIds = userIds[:len(userIds)-1]
				break
			}
		}
	}
	roomMembers, err := s.members.Users(room)
	if err != nil {
		return err
	}
	s.asyncEventSink.Send(roomMembers, state)
	return nil
}

func (s *typingStream) Typing(room types.RoomId) ([]types.UserId, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	state := s.states[room]
	if state == nil {
		return []types.UserId{}, nil
	}
	return state.event.Content.UserIds, nil
}

func (s *typingStream) Max() uint64 {
	return atomic.LoadUint64(&s.max)
}

// ignores user, userSet, and limit
func (s *typingStream) Range(
	_ *types.UserId,
	userSet map[types.UserId]struct{},
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit uint,
) ([]types.IndexedEvent, types.Error) {
	var result []types.IndexedEvent
	if len(roomSet) == 0 || from >= to {
		return result, nil
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	result = make([]types.IndexedEvent, 0, len(roomSet))
	for room := range roomSet {
		state := s.states[room]
		if state.index >= from && state.index < to {
			result = append(result, state)
		}
	}
	return result, nil
}
