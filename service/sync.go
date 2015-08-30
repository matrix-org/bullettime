package service

import (
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

func NewSyncService(
	messageSource interfaces.IndexedEventSource,
	presenceSource interfaces.IndexedEventSource,
	typingSource interfaces.IndexedEventSource,
	rooms interfaces.RoomStore,
	membershipStore interfaces.MembershipStore,
) (interfaces.SyncService, error) {
	return &syncService{
		messageSource,
		presenceSource,
		typingSource,
		rooms,
		membershipStore,
	}, nil
}

type syncService struct {
	messageSource   interfaces.IndexedEventSource
	presenceSource  interfaces.IndexedEventSource
	typingSource    interfaces.IndexedEventSource
	rooms           interfaces.RoomStore
	membershipStore interfaces.MembershipStore
}

func indexedToEvents(indexed []types.IndexedEvent) []types.Event {
	events := make([]types.Event, len(indexed))
	for i, indexedEvent := range indexed {
		events[i] = indexedEvent.Event()
	}
	return events
}

func (s syncService) FullSync(user types.UserId, limit uint) (*types.InitialSync, types.Error) {
	maxMessage := s.messageSource.Max()
	maxPresence := s.presenceSource.Max()
	maxTyping := s.typingSource.Max()

	userSet, err := s.membershipStore.Peers(user)
	if err != nil {
		return nil, err
	}
	indexedPresences, err := s.presenceSource.Range(user, userSet, nil, 0, maxPresence, limit)
	if err != nil {
		return nil, err
	}
	presences := indexedToEvents(indexedPresences)

	rooms, err := s.membershipStore.Rooms(user)
	summaries := make([]types.RoomSummary, len(rooms))
	if err != nil {
		return nil, err
	}
	for i, room := range rooms {
		if err := s.roomSummary(&summaries[i], user, room, maxMessage, limit); err != nil {
			return nil, err
		}
	}

	end := types.NewStreamToken(maxMessage, maxPresence, maxTyping)
	initialSync := types.InitialSync{end, presences, summaries}

	return &initialSync, nil
}

func (s syncService) RoomSync(user types.UserId, room types.RoomId, limit uint) (*types.RoomInitialSync, types.Error) {
	maxMessage := s.messageSource.Max()
	maxPresence := s.presenceSource.Max()

	userSet := map[types.UserId]struct{}{}
	users, err := s.membershipStore.Users(room)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userSet[user] = struct{}{}
	}

	indexedPresences, err := s.presenceSource.Range(user, userSet, nil, 0, maxPresence, limit)
	if err != nil {
		return nil, err
	}
	presences := indexedToEvents(indexedPresences)

	sync := types.RoomInitialSync{
		Presence: presences,
	}

	if err := s.roomSummary(&sync.RoomSummary, user, room, maxMessage, limit); err != nil {
		return nil, err
	}
	return &sync, nil
}

func (s syncService) roomSummary(
	summary *types.RoomSummary,
	user types.UserId,
	room types.RoomId,
	maxMessage uint64,
	limit uint,
) types.Error {
	roomSet := map[types.RoomId]struct{}{
		room: struct{}{},
	}
	messages, err := s.messageSource.Range(user, nil, roomSet, maxMessage, 0, limit)
	if err != nil {
		return err
	}
	states, err := s.rooms.EntireRoomState(room)
	if err != nil {
		return err
	}
	membershipState, err := s.rooms.RoomState(room, types.EventTypeMembership, user.String())
	if err != nil {
		return err
	}
	membership := membershipState.Content.(types.MembershipEventContent).Membership
	joinRuleState, err := s.rooms.RoomState(room, types.EventTypeJoinRules, user.String())
	if err != nil {
		return err
	}
	joinRule := joinRuleState.Content.(types.JoinRulesEventContent).JoinRule
	visibility := joinRule.ToVisibility()
	summary.Membership = membership
	summary.RoomId = room
	summary.Messages = indexedToEvents(messages)
	summary.State = states
	summary.Visibility = visibility
	return nil
}
