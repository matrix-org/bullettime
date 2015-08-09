package events

import (
	"container/list"
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

type indexedMessage struct {
	event types.Message
	index uint64
}

type messageSource struct {
	lock      sync.RWMutex
	list      *list.List
	byId      map[types.EventId]indexedMessage
	byIndex   []*indexedMessage
	max       uint64
	members   interfaces.MembershipStore
	eventSink interfaces.UserEventSink
}

func NewMessageSource(
	members interfaces.MembershipStore,
	eventSink interfaces.UserEventSink,
) (*messageSource, error) {
	return &messageSource{
		list:      list.New(),
		byId:      map[types.EventId]indexedMessage{},
		byIndex:   []*indexedMessage{},
		members:   members,
		eventSink: eventSink,
	}, nil
}

func (s *messageSource) Send(event *types.Message) (uint64, types.Error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	index := atomic.AddUint64(&s.max, 1) - 1
	item := indexedMessage{*event, index}

	if currentItem, ok := s.byId[event.EventId]; ok {
		s.byIndex[currentItem.index] = nil
	}
	s.byIndex = append(s.byIndex, &item)
	s.byId[event.EventId] = item

	users, err := s.members.Users(event.RoomId)
	if err != nil {
		return 0, nil
	}
	if extraUser := extraUserForEvent(event); extraUser != nil {
		l := len(users)
		allUsers := make([]types.UserId, l+1)
		copy(allUsers, users)
		allUsers[l] = *extraUser
		users = allUsers
	}
	s.eventSink.Send(users, event, index)
	return index, nil
}

func extraUserForEvent(event *types.Message) *types.UserId {
	if event.EventType == types.EventTypeMembership {
		switch event.Content.(types.MembershipEventContent).Membership {
		case types.MembershipInvited:
			return &event.UserId
		case types.MembershipKnocking:
			return &event.UserId
		case types.MembershipBanned:
			return &event.UserId
		}
	}
	return nil
}

func (s *messageSource) Event(eventId types.EventId) (types.Message, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.byId[eventId].event, nil
}

func (s *messageSource) Range(
	user types.UserId,
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit int,
) ([]types.Event, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	result := make([]types.Event, 0, limit)

	if from == to {
		return result, nil
	}
	max := atomic.LoadUint64(&s.max)
	if to < from && from >= max {
		from = max - 1
	}
	i := from
	for len(result) < limit && i < max && i >= 0 {
		item := s.byIndex[i]
		if item != nil {
			event := item.event
			_, ok := roomSet[event.RoomId]
			if ok {
				result = append(result, &event)
			} else if extra := extraUserForEvent(&event); extra != nil && *extra == user {
				result = append(result, &event)
			}
		}
		if to > from {
			i += 1
			if i >= to {
				break
			}
		} else {
			if i == 0 {
				break
			}
			i -= 1
			if i <= to {
				break
			}
		}
	}
	return result, nil
}

func (s *messageSource) Max() (index uint64, err types.Error) {
	return atomic.LoadUint64(&s.max), nil
}
