package events

import (
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"

	"github.com/Rugvip/bullettime/types"
)

type presenceStream struct {
	lock           sync.RWMutex
	events         map[types.UserId]indexedPresenceEvent
	max            uint64
	members        interfaces.MembershipStore
	asyncEventSink interfaces.AsyncEventSink
}

type indexedPresenceEvent struct {
	event types.PresenceEvent
	index uint64
}

func (m *indexedPresenceEvent) Event() types.Event {
	return &m.event
}

func (s *indexedPresenceEvent) Index() uint64 {
	return s.index
}

type updateFunc func(*types.User)

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

func (s *presenceStream) SetUserProfile(userId types.UserId, profile types.UserProfile) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *types.User) {
		user.UserProfile = profile
	})
}

func (s *presenceStream) SetUserPresence(userId types.UserId, presence types.UserPresence) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *types.User) {
		user.UserPresence = presence
	})
}

func (s *presenceStream) update(userId types.UserId, updateFunc updateFunc) (types.IndexedEvent, types.Error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	indexed, existed := s.events[userId]
	if !existed {
		indexed.event.Content.UserId = userId
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

func (s *presenceStream) User(user types.UserId) (types.User, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if indexed, ok := s.events[user]; ok {
		return indexed.event.Content, nil
	}
	return types.User{}, nil
}

func (s *presenceStream) Max() uint64 {
	return atomic.LoadUint64(&s.max)
}

// Ignores user, roomSet, and limit
func (s *presenceStream) Range(
	user types.UserId,
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
	result = make([]types.IndexedEvent, len(userSet))
	for user := range userSet {
		event := s.events[user]
		if event.index >= from && event.index < to {
			result = append(result, &event)
		}
	}
	return result, nil
}
