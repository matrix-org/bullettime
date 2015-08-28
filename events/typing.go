package events

import (
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"

	"github.com/Rugvip/bullettime/types"
)

type typingSource struct {
	lock      sync.RWMutex
	states    map[types.RoomId]*indexedTypingState
	max       uint64
	members   interfaces.MembershipStore
	eventSink interfaces.UserEventSink
}

type indexedTypingState struct {
	index uint64
	event types.TypingEvent
}

func NewTypingSource(
	members interfaces.MembershipStore,
	eventSink interfaces.UserEventSink,
) (interfaces.TypingStream, error) {
	return &typingSource{
		states:    map[types.RoomId]*indexedTypingState{},
		members:   members,
		eventSink: eventSink,
	}, nil
}

func (s *typingSource) SetTyping(room types.RoomId, user types.UserId, typing bool) types.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	state := s.states[room]
	index := atomic.AddUint64(&s.max, 1) - 1
	if state == nil {
		state = &indexedTypingState{index: index}
		state.event.RoomId = room
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
	s.eventSink.Send(roomMembers, &state.event, index)
	return nil
}

func (s *typingSource) Typing(room types.RoomId) ([]types.UserId, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	state := s.states[room]
	if state == nil {
		return []types.UserId{}, nil
	}
	return state.event.Content.UserIds, nil
}

func (s *typingSource) Max() (uint64, types.Error) {
	return atomic.LoadUint64(&s.max), nil
}

// ignores user, userSet, and limit
func (s *typingSource) Range(
	user types.UserId,
	userSet map[types.UserId]struct{},
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit int,
) ([]types.Event, types.Error) {
	var result []types.Event
	if len(roomSet) == 0 || from >= to {
		return result, nil
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	result = make([]types.Event, len(roomSet))
	for room := range roomSet {
		state := s.states[room]
		if state.index >= from && state.index < to {
			result = append(result, &state.event)
		}
	}
	return result, nil
}
