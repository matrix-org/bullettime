// Copyright 2015 OpenMarket Ltd
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
	indexedPresences, err := s.presenceSource.Range(&user, userSet, nil, 0, maxPresence, limit)
	if err != nil {
		return nil, err
	}
	presences := indexedToEvents(indexedPresences)

	rooms, err := s.membershipStore.Rooms(user)
	summaries := make([]types.RoomSummary, len(rooms))
	if err != nil {
		return nil, err
	}
	end := types.NewStreamToken(maxMessage, maxPresence, maxTyping)

	for i, room := range rooms {
		if err := s.roomSummary(&summaries[i], user, room, end, limit); err != nil {
			return nil, err
		}
	}

	initialSync := types.InitialSync{end, presences, summaries}

	return &initialSync, nil
}

func (s syncService) RoomSync(user types.UserId, room types.RoomId, limit uint) (*types.RoomInitialSync, types.Error) {
	maxMessage := s.messageSource.Max()
	maxPresence := s.presenceSource.Max()
	maxTyping := s.typingSource.Max()

	userSet := map[types.UserId]struct{}{}
	users, err := s.membershipStore.Users(room)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userSet[user] = struct{}{}
	}

	indexedPresences, err := s.presenceSource.Range(&user, userSet, nil, 0, maxPresence, limit)
	if err != nil {
		return nil, err
	}
	presences := indexedToEvents(indexedPresences)

	sync := types.RoomInitialSync{
		Presence: presences,
	}

	end := types.NewStreamToken(maxMessage, maxPresence, maxTyping)
	if err := s.roomSummary(&sync.RoomSummary, user, room, end, limit); err != nil {
		return nil, err
	}
	return &sync, nil
}

func (s syncService) roomSummary(
	summary *types.RoomSummary,
	user types.UserId,
	room types.RoomId,
	end types.StreamToken,
	limit uint,
) types.Error {
	roomSet := map[types.RoomId]struct{}{
		room: struct{}{},
	}
	messages, err := s.messageSource.Range(nil, nil, roomSet, end.MessageIndex, 0, limit)
	if err != nil {
		return err
	}
	startIndex := end.MessageIndex
	if len(messages) > 0 {
		startIndex = messages[0].Index()
	}
	start := types.NewStreamToken(startIndex, end.PresenceIndex, end.TypingIndex)
	eventRange := types.NewEventStreamRange(indexedToEvents(messages), start, end)
	states, err := s.rooms.EntireRoomState(room)
	if err != nil {
		return err
	}
	membershipState, err := s.rooms.RoomState(room, types.EventTypeMembership, user.String())
	if err != nil {
		return err
	}
	membership := membershipState.Content.(*types.MembershipEventContent).Membership
	joinRuleState, err := s.rooms.RoomState(room, types.EventTypeJoinRules, "")
	if err != nil {
		return err
	}
	joinRule := joinRuleState.Content.(*types.JoinRulesEventContent).JoinRule
	visibility := joinRule.ToVisibility()
	summary.Membership = membership
	summary.RoomId = room
	summary.Messages = eventRange
	summary.State = states
	summary.Visibility = visibility
	return nil
}
