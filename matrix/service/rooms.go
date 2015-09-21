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

package service

import (
	"fmt"
	"log"
	"reflect"
	"time"

	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	"github.com/matrix-org/bullettime/matrix/types"
	"github.com/matrix-org/bullettime/utils"
)

func CreateRoomService(
	roomStore interfaces.RoomStore,
	aliasStore interfaces.AliasStore,
	memberStore interfaces.MembershipStore,
	eventSink interfaces.EventSink,
	profileProvider interfaces.ProfileProvider,
	typingSink interfaces.TypingEventSink,
	typingProvider interfaces.TypingProvider,
) (interfaces.RoomService, error) {
	return roomService{
		roomStore,
		aliasStore,
		memberStore,
		eventSink,
		profileProvider,
		typingSink,
		typingProvider,
	}, nil
}

type roomService struct {
	rooms           interfaces.RoomStore
	aliases         interfaces.AliasStore
	members         interfaces.MembershipStore
	eventSink       interfaces.EventSink
	profileProvider interfaces.ProfileProvider
	typingSink      interfaces.TypingEventSink
	typingProvider  interfaces.TypingProvider
}

func (s roomService) RoomExists(id ct.RoomId, caller ct.UserId) types.Error {
	exists, err := s.rooms.RoomExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return types.NotFoundError("room '" + id.String() + "' doesn't exist")
	}
	return nil
}

func (s roomService) LookupAlias(alias ct.Alias) (ct.RoomId, types.Error) {
	room, err := s.aliases.Room(alias)
	if err != nil {
		return ct.RoomId{}, err
	}
	if room == nil {
		return ct.RoomId{}, types.NotFoundError("room alias '" + alias.String() + "' doesn't exist")
	}
	return *room, nil
}

func (s roomService) CreateRoom(
	domain string,
	creator ct.UserId,
	desc *types.RoomDescription,
) (ct.RoomId, *ct.Alias, types.Error) {
	var alias *ct.Alias
	id := ct.NewRoomId(utils.RandomString(16), domain)
	if desc.Alias != nil {
		a := ct.NewAlias(*desc.Alias, domain)
		err := s.aliases.AddAlias(a, id)
		if err != nil {
			return ct.RoomId{}, nil, err
		}
		alias = &a
	}
	exists, err := s.rooms.CreateRoom(id)
	if exists {
		return ct.RoomId{}, nil, types.RoomInUseError("room '" + id.String() + "' already exists")
	}
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	s.members.AddMember(id, creator)

	_, err = s.sendMessage(id, creator, &types.CreateEventContent{creator})
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	_, err = s.setState(id, creator, &types.CreateEventContent{creator}, "")
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	profile, err := s.profileProvider.Profile(creator)
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	membership := &types.MembershipEventContent{&profile, types.MembershipMember}
	_, err = s.setState(id, creator, membership, creator.String())
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	_, err = s.setState(id, creator, types.DefaultPowerLevels(creator), "")
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	joinRuleContent := types.JoinRulesEventContent{desc.Visibility.ToJoinRule()}
	_, err = s.setState(id, creator, &joinRuleContent, "")
	if err != nil {
		return ct.RoomId{}, nil, err
	}
	if alias != nil {
		_, err = s.setState(id, creator, &types.AliasesEventContent{[]ct.Alias{*alias}}, "")
		if err != nil {
			return ct.RoomId{}, nil, err
		}
	}
	if desc.Name != nil {
		_, err = s.setState(id, creator, &types.NameEventContent{*desc.Name}, "")
		if err != nil {
			return ct.RoomId{}, nil, err
		}
	}
	if desc.Topic != nil {
		_, err = s.setState(id, creator, &types.TopicEventContent{*desc.Topic}, "")
		if err != nil {
			return ct.RoomId{}, nil, err
		}
	}
	for _, invited := range desc.Invited {
		membership := types.MembershipEventContent{nil, types.MembershipInvited}
		_, err = s.setState(id, creator, &membership, invited.String())
		if err != nil {
			return ct.RoomId{}, nil, err
		}
	}
	return id, alias, nil
}

var disallowedMessageTypes map[string]struct{} = map[string]struct{}{
	types.EventTypeName:        struct{}{},
	types.EventTypeTopic:       struct{}{},
	types.EventTypeJoinRules:   struct{}{},
	types.EventTypePowerLevels: struct{}{},
	types.EventTypeCreate:      struct{}{},
	types.EventTypeAliases:     struct{}{},
	types.EventTypeMembership:  struct{}{},
}

func (s roomService) AddMessage(
	room ct.RoomId,
	caller ct.UserId,
	content ct.TypedContent,
) (*types.Message, types.Error) {
	eventType := content.GetEventType()
	if _, ok := disallowedMessageTypes[eventType]; ok {
		return nil, types.ForbiddenError("sending a message event of the type " + eventType + " is not permitted")
	}

	err := s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
		if eventLevel, ok := pl.Events[eventType]; ok {
			return eventLevel
		}
		return pl.EventDefault
	})
	if err != nil {
		return nil, err
	}

	return s.sendMessage(room, caller, content)
}

func (s roomService) State(
	room ct.RoomId,
	caller ct.UserId,
	eventType, stateKey string,
) (*types.State, types.Error) {
	membership, err := s.userMembership(room, caller)
	if err != nil {
		return nil, err
	}
	if membership != types.MembershipMember {
		return nil, types.ForbiddenError("cannot read room state, not a member")
	}
	state, err := s.rooms.RoomState(room, eventType, stateKey)
	if err != nil {
		return nil, err
	}
	return state, err
}

func (s roomService) SetState(
	room ct.RoomId,
	caller ct.UserId,
	content ct.TypedContent,
	stateKey string,
) (*types.State, types.Error) {
	userIdStateKey, parseErr := ct.ParseUserId(stateKey)
	isUserIdStateKey := parseErr == nil

	eventType := content.GetEventType()
	switch eventType {
	case types.EventTypeName:
		if stateKey != "" {
			return nil, types.ForbiddenError("state key must be empty for state " + eventType)
		}
	case types.EventTypeTopic:
		if stateKey != "" {
			return nil, types.ForbiddenError("state key must be empty for state " + eventType)
		}
	case types.EventTypeJoinRules:
		if stateKey != "" {
			return nil, types.ForbiddenError("state key must be empty for state " + eventType)
		}
	case types.EventTypePowerLevels:
		if stateKey != "" {
			return nil, types.ForbiddenError("state key must be empty for state " + eventType)
		}
	case types.EventTypeCreate:
		return nil, types.ForbiddenError("cannot set state " + eventType)

	case types.EventTypeAliases:
		return nil, types.ForbiddenError("cannot set state " + eventType)

	case types.EventTypeMembership:
		membership, ok := content.(*types.MembershipEventContent)
		if !ok || membership == nil {
			panic("expected membership event content, got " + reflect.TypeOf(content).String())
		}
		if !isUserIdStateKey {
			return nil, types.ForbiddenError("state key must be a user id for state " + eventType)
		}
		state, err := s.doMembershipChange(room, caller, userIdStateKey, membership)
		if err != nil {
			return nil, err
		}
		return state, nil
	}
	if isUserIdStateKey && userIdStateKey != caller {
		return nil, types.ForbiddenError("cannot set the state of another user")
	}

	existing, err := s.rooms.RoomState(room, eventType, stateKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		err := s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
			return pl.CreateState
		})
		if err != nil {
			return nil, err
		}
	}
	err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
		if eventLevel, ok := pl.Events[eventType]; ok {
			return eventLevel
		}
		return pl.EventDefault
	})
	if err != nil {
		return nil, err
	}
	return s.setState(room, caller, content, stateKey)
}

func (s roomService) setState(
	room ct.RoomId,
	user ct.UserId,
	content ct.TypedContent,
	stateKey string,
) (*types.State, types.Error) {
	log.Printf("Setting state: %#v, %#v, %#v, %#v", room, user, content, stateKey)
	state, err := s.rooms.SetRoomState(room, user, content, stateKey)
	if err != nil {
		return nil, err
	}
	_, err = s.eventSink.Send(state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s roomService) sendMessage(
	room ct.RoomId,
	user ct.UserId,
	content ct.TypedContent,
) (*types.Message, types.Error) {
	log.Printf("Sending message: %#v, %#v, %#v, %#v", room, user, content)

	message := new(types.Message)
	message.EventId = ct.DeriveEventId(utils.RandomString(16), ct.Id(user))
	message.RoomId = room
	message.UserId = user
	message.EventType = content.GetEventType()
	message.Timestamp = ct.Timestamp{time.Now()}
	message.Content = content

	_, err := s.eventSink.Send(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (s roomService) doMembershipChange(
	room ct.RoomId,
	caller ct.UserId,
	user ct.UserId,
	membership *types.MembershipEventContent,
) (*types.State, types.Error) {
	log.Printf("attempting membership change of %s in %s to %s, by %s", user, room, membership.Membership, caller)
	currentMembership, err := s.userMembership(room, user)
	if err != nil {
		return nil, err
	}
	if currentMembership == membership.Membership {
		return nil, types.ForbiddenError("membership change was a no-op")
	}
	membership.UserProfile = nil

	switch membership.Membership {
	case types.MembershipNone:
		if currentMembership != types.MembershipBanned {
			return nil, types.BadJsonError("invalid or missing membership in membership change")
		}
		err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
		if user == caller {
			return nil, types.ForbiddenError("cannot remove a ban from self")
		}

	case types.MembershipInvited:
		if currentMembership != types.MembershipNone {
			return nil, types.ForbiddenError("could not invite user to room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := s.allowsJoinRule(room, types.JoinRuleInvite)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, types.ForbiddenError("room does not allow join method: " + types.JoinRuleInvite.String())
		}
		err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
			return pl.Invite
		})
		if err != nil {
			return nil, err
		}

	case types.MembershipMember:
		switch currentMembership {
		case types.MembershipNone:
			ok, err := s.allowsJoinRule(room, types.JoinRulePublic)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, types.ForbiddenError("room does not allow join method: " + types.JoinRuleInvite.String())
			}
		case types.MembershipInvited:
			if user != caller {
				return nil, types.ForbiddenError("cannot force other users to join the room")
			}
		case types.MembershipKnocking:
			if user == caller {
				return nil, types.ForbiddenError("cannot let yourself in after knocking")
			}
			err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
				return pl.Invite
			})
			if err != nil {
				return nil, err
			}
		case types.MembershipBanned:
			if user == caller {
				return nil, types.ForbiddenError("you are banned from that room")
			} else {
				return nil, types.ForbiddenError("that user is banned from this room")
			}
		}
		profile, err := s.profileProvider.Profile(caller)
		if err != nil {
			return nil, err
		}
		membership.UserProfile = &profile

	case types.MembershipKnocking:
		if user != caller {
			return nil, types.ForbiddenError("cannot force other users to knock")
		}
		if currentMembership != types.MembershipNone {
			return nil, types.ForbiddenError("could not knock on room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := s.allowsJoinRule(room, types.JoinRuleKnock)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, types.ForbiddenError("room does not allow join method: " + types.JoinRuleKnock.String())
		}

	case types.MembershipLeaving:
		if currentMembership == types.MembershipNone {
			return nil, types.ForbiddenError("tried to leave a room without current membership")
		}
		if currentMembership == types.MembershipBanned {
			return nil, types.ForbiddenError("tried to leave room with current membership '" + types.MembershipBanned.String() + "'")
		}
		if user != caller {
			err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
				return pl.Kick
			})
			if err != nil {
				return nil, err
			}
		}

	case types.MembershipBanned:
		if user == caller {
			return nil, types.ForbiddenError("cannot ban self")
		}
		err = s.testPowerLevel(room, caller, func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
	}
	if membership.Membership == types.MembershipMember {
		if err := s.members.AddMember(room, user); err != nil {
			return nil, err
		}
	} else if currentMembership == types.MembershipMember {
		if err := s.members.RemoveMember(room, user); err != nil {
			return nil, err
		}
	}
	return s.setState(room, caller, membership, user.String())
}

func (s roomService) testPowerLevel(
	room ct.RoomId,
	user ct.UserId,
	powerLevelFunc func(*types.PowerLevelsEventContent) int,
) types.Error {
	powerLevels, err := s.powerLevels(room)
	if err != nil {
		return err
	}
	userPowerLevel, err := s.userPowerLevel(room, user)
	if err != nil {
		return err
	}
	requiredPowerLevel := powerLevelFunc(powerLevels)
	if userPowerLevel < requiredPowerLevel {
		msg := fmt.Sprintf("not enough power level to perform action (%d < %d)", userPowerLevel, requiredPowerLevel)
		return types.ForbiddenError(msg)
	}
	return nil
}

func (s roomService) userMembership(room ct.RoomId, user ct.UserId) (types.Membership, types.Error) {
	state, err := s.rooms.RoomState(room, types.EventTypeMembership, user.String())
	if err != nil {
		return types.MembershipNone, err
	}
	if state == nil {
		return types.MembershipNone, nil
	}
	membership, ok := state.Content.(*types.MembershipEventContent)
	if !ok {
		panic("invalid membership content, was " + reflect.TypeOf(state.Content).String())
	}
	return membership.Membership, nil
}

func (s roomService) allowsJoinRule(room ct.RoomId, joinRule types.JoinRule) (bool, types.Error) {
	state, err := s.rooms.RoomState(room, types.EventTypeJoinRules, "")
	if err != nil {
		return false, err
	}
	if state == nil {
		panic("room power levels are invalid or missing: " + room.String())
	}
	joinRules, ok := state.Content.(*types.JoinRulesEventContent)
	if !ok {
		panic("invalid join rule content, was " + reflect.TypeOf(state.Content).String())
	}
	if joinRules.JoinRule != joinRule {
		return false, types.ForbiddenError("room does not allow join rule: " + joinRule.String())
	}
	return true, nil
}

func (s roomService) powerLevels(room ct.RoomId) (*types.PowerLevelsEventContent, types.Error) {
	state, err := s.rooms.RoomState(room, types.EventTypePowerLevels, "")
	if err != nil {
		return nil, err
	}
	if state == nil {
		panic("room power levels are invalid or missing: " + room.String())
	}
	powerLevels, ok := state.Content.(*types.PowerLevelsEventContent)
	if !ok {
		panic("invalid power level content, was " + reflect.TypeOf(state.Content).String())
	}
	return powerLevels, nil
}

func (s roomService) userPowerLevel(room ct.RoomId, user ct.UserId) (int, types.Error) {
	powerLevels, err := s.powerLevels(room)
	if err != nil {
		return 0, err
	}
	if userLevel, ok := powerLevels.Users[user.String()]; ok {
		return userLevel, nil
	}
	return powerLevels.UserDefault, nil
}

func (s roomService) eventPowerLevel(room ct.RoomId, eventType string) (int, types.Error) {
	powerLevels, err := s.powerLevels(room)
	if err != nil {
		return 0, err
	}
	if eventLevel, ok := powerLevels.Events[eventType]; ok {
		return eventLevel, nil
	}
	return powerLevels.EventDefault, nil
}
