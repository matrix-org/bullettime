package events

import (
	"container/list"
	"log"
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

type indexedEvent struct {
	event types.Event
	index uint64
}

func (m *indexedEvent) Event() types.Event {
	return m.event
}

func (m *indexedEvent) Index() uint64 {
	return m.index
}

type messageStream struct {
	lock           sync.RWMutex
	list           *list.List
	byId           map[types.EventId]indexedEvent
	byIndex        []*indexedEvent
	max            uint64
	members        interfaces.MembershipStore
	asyncEventSink interfaces.AsyncEventSink
}

func NewMessageStream(
	members interfaces.MembershipStore,
	asyncEventSink interfaces.AsyncEventSink,
) (interfaces.EventStream, error) {
	return &messageStream{
		list:           list.New(),
		byId:           map[types.EventId]indexedEvent{},
		byIndex:        []*indexedEvent{},
		members:        members,
		asyncEventSink: asyncEventSink,
	}, nil
}

func (s *messageStream) Send(event types.Event) (uint64, types.Error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	index := atomic.AddUint64(&s.max, 1) - 1
	indexed := indexedEvent{event, index}

	if currentItem, ok := s.byId[*event.GetEventId()]; ok {
		s.byIndex[currentItem.index] = nil
	}
	s.byIndex = append(s.byIndex, &indexed)
	s.byId[*event.GetEventId()] = indexed

	users, err := s.members.Users(*event.GetRoomId())
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
	s.asyncEventSink.Send(users, &indexed)
	return index, nil
}

func extraUserForEvent(event types.Event) *types.UserId {
	if event.GetEventType() == types.EventTypeMembership {
		membership := event.GetContent().(*types.MembershipEventContent).Membership
		isInvited := membership == types.MembershipInvited
		isKnocking := membership == types.MembershipKnocking
		isBanned := membership == types.MembershipBanned
		if isInvited || isKnocking || isBanned {
			state, ok := event.(*types.State)
			if !ok {
				log.Println("membership event was not a state event:", event)
				return nil
			}
			user, err := types.ParseUserId(state.StateKey)
			if err != nil {
				log.Println("failed to parse user id state key:", state.StateKey)
				return nil
			}
			return &user
		}
	}
	return nil
}

func (s *messageStream) Event(
	user types.UserId,
	eventId types.EventId,
) (types.Event, types.Error) {
	s.lock.RLock()
	indexed := s.byId[eventId]
	s.lock.RUnlock()
	extraUser := extraUserForEvent(indexed.event)
	if extraUser != nil && *extraUser == user {
		return indexed.event, nil
	}
	rooms, err := s.members.Rooms(user)
	if err != nil {
		return nil, err
	}
	for _, room := range rooms {
		if room == *indexed.event.GetRoomId() {
			return indexed.event, nil
		}
	}
	return nil, nil
}

// ignores userSet
func (s *messageStream) Range(
	user types.UserId,
	userSet map[types.UserId]struct{},
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit uint,
) ([]types.IndexedEvent, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	result := make([]types.IndexedEvent, 0, limit)
	reverse := to < from

	max := atomic.LoadUint64(&s.max)
	if reverse {
		if from >= max {
			from = max
		}
		from -= 1
		if from < to {
			return result, nil
		}
	} else {
		if from == to {
			return result, nil
		}
	}
	i := from
	for uint(len(result)) < limit && i < max {
		indexed := s.byIndex[i]
		if indexed != nil {
			_, ok := roomSet[*indexed.Event().GetRoomId()]
			if ok {
				result = append(result, indexed)
			} else if extra := extraUserForEvent(indexed.event); extra != nil && *extra == user {
				result = append(result, indexed)
			}
		}
		if reverse {
			i -= 1
			if i < to {
				break
			}
		} else {
			i += 1
			if i >= to {
				break
			}
		}
	}
	return result, nil
}

func (s *messageStream) Max() uint64 {
	return atomic.LoadUint64(&s.max)
}
