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

type presenceStream struct {
	lock           sync.RWMutex
	events         map[types.UserId]indexedPresenceEvent
	max            uint64
	members        interfaces.MembershipStore
	asyncEventSink interfaces.AsyncEventSink
}

type indexedPresenceEvent struct {
	event matrixTypes.PresenceEvent
	index uint64
}

func (m *indexedPresenceEvent) Event() types.Event {
	return &m.event
}

func (s *indexedPresenceEvent) Index() uint64 {
	return s.index
}

type updateFunc func(*matrixTypes.User)

func NewPresenceStream(
	members interfaces.MembershipStore,
	asyncEventSink interfaces.AsyncEventSink,
) (interfaces.PresenceStream, error) {
	return &presenceStream{
		events:         map[types.UserId]indexedPresenceEvent{},
		members:        members,
		asyncEventSink: asyncEventSink,
	}, nil
}

func (s *presenceStream) SetUserProfile(userId types.UserId, profile matrixTypes.UserProfile) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *matrixTypes.User) {
		user.UserProfile = profile
	})
}

func (s *presenceStream) SetUserStatus(userId types.UserId, status matrixTypes.UserStatus) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *matrixTypes.User) {
		user.UserStatus = status
	})
}

func (s *presenceStream) update(userId types.UserId, updateFunc updateFunc) (types.IndexedEvent, types.Error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	indexed, existed := s.events[userId]
	if !existed {
		indexed.event.Content.UserId = userId
		indexed.event.EventType = matrixTypes.EventTypePresence
	}
	updateFunc(&indexed.event.Content)
	index := atomic.AddUint64(&s.max, 1) - 1
	indexed.index = index
	s.events[userId] = indexed
	peerSet, err := s.members.Peers(userId)
	if err != nil {
		return nil, err
	}
	peers := make([]types.UserId, len(peerSet))
	for peer := range peerSet {
		peers = append(peers, peer)
	}
	s.asyncEventSink.Send(peers, &indexed)
	return &indexed, nil
}

func (s *presenceStream) Profile(user types.UserId) (matrixTypes.UserProfile, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if indexed, ok := s.events[user]; ok {
		return indexed.event.Content.UserProfile, nil
	}
	return matrixTypes.UserProfile{}, nil
}

func (s *presenceStream) Status(user types.UserId) (matrixTypes.UserStatus, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if indexed, ok := s.events[user]; ok {
		return indexed.event.Content.UserStatus, nil
	}
	return matrixTypes.UserStatus{}, nil
}

func (s *presenceStream) Max() uint64 {
	return atomic.LoadUint64(&s.max)
}

// Ignores user, roomSet, and limit
func (s *presenceStream) Range(
	_ *types.UserId,
	userSet map[types.UserId]struct{},
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit uint,
) ([]types.IndexedEvent, types.Error) {
	var result []types.IndexedEvent
	if len(userSet) == 0 || from >= to {
		return result, nil
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	result = make([]types.IndexedEvent, 0, len(userSet))
	for user := range userSet {
		event := s.events[user]
		if event.index >= from && event.index < to {
			result = append(result, &event)
		}
	}
	return result, nil
}
