package service

import (
	"fmt"
	"reflect"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

func CreateRoomService(
	roomStore interfaces.RoomStore,
	aliasStore interfaces.AliasStore,
	memberStore interfaces.MembershipStore,
	userStateStore interfaces.UserStateStore,
	eventSink interfaces.EventSink,
	typingSink interfaces.TypingEventSink,
	typingStore interfaces.TypingStore,
) (interfaces.RoomService, error) {
	return roomService{
		roomStore,
		aliasStore,
		memberStore,
		userStateStore,
		eventSink,
		typingSink,
		typingStore,
	}, nil
}

type roomService struct {
	rooms          interfaces.RoomStore
	aliases        interfaces.AliasStore
	members        interfaces.MembershipStore
	userStateStore interfaces.UserStateStore
	eventSink      interfaces.EventSink
	typingSink     interfaces.TypingEventSink
	typingStore    interfaces.TypingStore
}

type Room struct {
	id      types.RoomId
	service roomService
}

func (r roomService) Room(id types.RoomId) (interfaces.Room, types.Error) {
	exists, err := r.rooms.RoomExists(id)
	if err != nil {
		return Room{}, err
	}
	if !exists {
		return nil, types.NotFoundError("room '" + id.String() + "' doesn't exist")
	}
	return Room{id, r}, nil
}

func (r roomService) CreateRoom(
	domain string,
	creator interfaces.User,
	desc *types.RoomDescription,
) (interfaces.Room, *types.Alias, types.Error) {
	var alias *types.Alias
	if desc.Alias != nil {
		a := types.NewAlias(*desc.Alias, domain)
		alias = &a
		if err := r.aliases.Reserve(*alias); err != nil {
			return nil, nil, err
		}
	}
	id, err := r.rooms.CreateRoom(domain)
	if err != nil {
		return nil, nil, err
	}
	if alias != nil {
		if err := r.aliases.Claim(*alias, id); err != nil {
			return nil, nil, err
		}
	}
	userId := creator.Id()
	r.members.AddMember(id, userId)
	_, err = r.setState(id, userId, &types.CreateEventContent{userId}, "")
	if err != nil {
		return nil, nil, err
	}
	profile, err := creator.Profile()
	if err != nil {
		return nil, nil, err
	}
	membership := &types.MembershipEventContent{&profile, types.MembershipMember}
	_, err = r.setState(id, userId, membership, userId.String())
	if err != nil {
		return nil, nil, err
	}
	_, err = r.setState(id, userId, types.DefaultPowerLevels(userId), "")
	if err != nil {
		return nil, nil, err
	}
	joinRuleContent := types.JoinRulesEventContent{desc.Visibility.ToJoinRule()}
	_, err = r.setState(id, userId, &joinRuleContent, "")
	if err != nil {
		return nil, nil, err
	}
	if alias != nil {
		_, err = r.setState(id, userId, &types.AliasesEventContent{[]types.Alias{*alias}}, "")
		if err != nil {
			return nil, nil, err
		}
	}
	if desc.Name != nil {
		_, err = r.setState(id, userId, &types.NameEventContent{*desc.Name}, "")
		if err != nil {
			return nil, nil, err
		}
	}
	if desc.Topic != nil {
		_, err = r.setState(id, userId, &types.TopicEventContent{*desc.Topic}, "")
		if err != nil {
			return nil, nil, err
		}
	}
	for _, invited := range desc.Invited {
		membership := types.MembershipEventContent{nil, types.MembershipInvited}
		_, err = r.setState(id, userId, &membership, invited.String())
		if err != nil {
			return nil, nil, err
		}
	}
	return Room{id, r}, alias, nil
}

func (s roomService) setState(
	roomId types.RoomId,
	userId types.UserId,
	content types.TypedContent,
	stateKey string,
) (*types.State, types.Error) {
	state, err := s.rooms.SetRoomState(roomId, userId, content, stateKey)
	if err != nil {
		return nil, err
	}
	_, err = s.eventSink.Send(state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (r Room) Id() types.RoomId {
	return r.id
}

func (r Room) AddMessage(user interfaces.User, content types.TypedContent) (*types.Message, types.Error) {
	return nil, nil
}

func (r Room) State(user interfaces.User, eventType, stateKey string) (*types.State, types.Error) {
	membership, err := r.userMembership(user.Id())
	if err != nil {
		return nil, err
	}
	if membership != types.MembershipMember {
		return nil, types.ForbiddenError("cannot read room state, not a member")
	}
	return r.service.rooms.RoomState(r.Id(), eventType, stateKey)
}

func (r Room) SetState(user interfaces.User, content types.TypedContent, stateKey string) (*types.State, types.Error) {
	userIdStateKey, err := types.ParseUserId(stateKey)
	isUserIdStateKey := err == nil

	eventType := content.EventType()
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
		return r.doMembershipChange(user, userIdStateKey, membership)
	}
	if isUserIdStateKey && userIdStateKey != user.Id() {
		return nil, types.ForbiddenError("cannot set the state of another user")
	}
	return nil, nil
}

func (r Room) doMembershipChange(by interfaces.User, userId types.UserId, membership *types.MembershipEventContent) (*types.State, types.Error) {
	currentMembership, err := r.userMembership(userId)
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
		err = r.testPowerLevel(by.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
		if userId == by.Id() {
			return nil, types.ForbiddenError("cannot remove a ban from self")
		}

	case types.MembershipInvited:
		if currentMembership != types.MembershipNone {
			return nil, types.ForbiddenError("could not invite user to room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := r.allowsJoinRule(types.JoinRuleInvite)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, types.ForbiddenError("room does not allow join method: " + types.JoinRuleInvite.String())
		}
		err = r.testPowerLevel(by.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Invite
		})
		if err != nil {
			return nil, err
		}

	case types.MembershipMember:
		if userId != by.Id() {
			return nil, types.ForbiddenError("cannot force other users to join the room")
		}
		profile, err := by.Profile()
		if err != nil {
			return nil, err
		}
		membership.UserProfile = &profile

	case types.MembershipKnocking:
		if userId != by.Id() {
			return nil, types.ForbiddenError("cannot force other users to knock")
		}
		if currentMembership != types.MembershipNone {
			return nil, types.ForbiddenError("could not knock on room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := r.allowsJoinRule(types.JoinRuleKnock)
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
		if userId != by.Id() {
			err = r.testPowerLevel(by.Id(), func(pl *types.PowerLevelsEventContent) int {
				return pl.Kick
			})
			if err != nil {
				return nil, err
			}
		}

	case types.MembershipBanned:
		if userId == by.Id() {
			return nil, types.ForbiddenError("cannot ban self")
		}
		err = r.testPowerLevel(by.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
	}
	if membership.Membership == types.MembershipMember {
		if err := r.service.members.AddMember(r.Id(), userId); err != nil {
			return nil, err
		}
	} else if currentMembership == types.MembershipMember {
		if err := r.service.members.RemoveMember(r.Id(), userId); err != nil {
			return nil, err
		}
	}
	return r.service.setState(r.Id(), by.Id(), membership, userId.String())
}

func (r Room) testPowerLevel(userId types.UserId, powerLevelFunc func(*types.PowerLevelsEventContent) int) types.Error {
	powerLevels, err := r.powerLevels()
	if err != nil {
		return err
	}
	userPowerLevel, err := r.userPowerLevel(userId)
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

func (r Room) userMembership(userId types.UserId) (types.Membership, types.Error) {
	state, err := r.service.rooms.RoomState(r.Id(), types.EventTypeMembership, userId.String())
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

func (r Room) allowsJoinRule(joinRule types.JoinRule) (bool, types.Error) {
	state, err := r.service.rooms.RoomState(r.Id(), types.EventTypeJoinRules, "")
	if err != nil {
		return false, err
	}
	if state == nil {
		panic("room power levels are invalid or missing: " + r.Id().String())
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

func (r Room) powerLevels() (*types.PowerLevelsEventContent, types.Error) {
	state, err := r.service.rooms.RoomState(r.Id(), types.EventTypePowerLevels, "")
	if err != nil {
		return nil, err
	}
	if state == nil {
		panic("room power levels are invalid or missing: " + r.Id().String())
	}
	powerLevels, ok := state.Content.(*types.PowerLevelsEventContent)
	if !ok {
		panic("invalid power level content, was " + reflect.TypeOf(state.Content).String())
	}
	return powerLevels, nil
}

func (r Room) userPowerLevel(userId types.UserId) (int, types.Error) {
	powerLevels, err := r.powerLevels()
	if err != nil {
		return 0, err
	}
	if userLevel, ok := powerLevels.Users[userId.String()]; ok {
		return userLevel, nil
	}
	return powerLevels.UserDefault, nil
}

func (r Room) eventPowerLevel(eventType string) (int, types.Error) {
	powerLevels, err := r.powerLevels()
	if err != nil {
		return 0, err
	}
	if eventLevel, ok := powerLevels.Events[eventType]; ok {
		return eventLevel, nil
	}
	return powerLevels.EventDefault, nil
}
