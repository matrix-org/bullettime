package service

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/types"
)

type Room struct {
	id types.RoomId
}

func GetRoom(id types.RoomId) (Room, error) {
	if err := db.RoomExists(id); err != nil {
		return Room{}, err
	}
	return Room{id: id}, nil
}

func CreateRoom(hostname string, creator User, desc *types.RoomDescription) (types.RoomId, *types.Alias, error) {
	var alias *types.Alias
	if desc.Alias != nil {
		a := types.NewAlias(*desc.Alias, hostname)
		alias = &a
	}
	id, err := db.CreateRoom(hostname, alias)
	if err != nil {
		return types.RoomId{}, nil, err
	}
	userId := creator.Id()
	_, err = db.SetRoomState(id, userId, &types.CreateEventContent{userId}, "")
	if err != nil {
		return types.RoomId{}, nil, err
	}
	profile, err := creator.GetProfile()
	if err != nil {
		return types.RoomId{}, nil, err
	}
	membership := &types.MembershipEventContent{&profile, types.MembershipMember}
	_, err = db.SetRoomState(id, userId, membership, userId.String())
	if err != nil {
		return types.RoomId{}, nil, err
	}
	_, err = db.SetRoomState(id, userId, types.DefaultPowerLevels(userId), "")
	if err != nil {
		return types.RoomId{}, nil, err
	}
	joinRuleContent := types.JoinRulesEventContent{desc.Visibility.ToJoinRule()}
	_, err = db.SetRoomState(id, userId, &joinRuleContent, "")
	if err != nil {
		return types.RoomId{}, nil, err
	}
	if alias != nil {
		_, err = db.SetRoomState(id, userId, &types.AliasesEventContent{[]types.Alias{*alias}}, "")
		if err != nil {
			return types.RoomId{}, nil, err
		}
	}
	if desc.Name != nil {
		_, err = db.SetRoomState(id, userId, &types.NameEventContent{*desc.Name}, "")
		if err != nil {
			return types.RoomId{}, nil, err
		}
	}
	if desc.Topic != nil {
		_, err = db.SetRoomState(id, userId, &types.TopicEventContent{*desc.Topic}, "")
		if err != nil {
			return types.RoomId{}, nil, err
		}
	}
	for _, invited := range desc.Invited {
		membership := types.MembershipEventContent{nil, types.MembershipInvited}
		_, err = db.SetRoomState(id, userId, &membership, invited.String())
		if err != nil {
			return types.RoomId{}, nil, err
		}
	}
	return id, alias, nil
}

func (r Room) Id() types.RoomId {
	return r.id
}

func (r Room) AddEvent(user User, content types.GenericContent) (*types.Event, error) {
	return nil, nil
}

func (r Room) SetState(user User, content types.TypedContent, stateKey string) (*types.State, error) {
	userIdStateKey, err := types.ParseUserId(stateKey)
	isUserIdStateKey := err == nil

	eventType := content.EventType()
	switch eventType {
	case types.EventTypeName:
		if stateKey != "" {
			return nil, errors.New("state key must be empty for state " + eventType)
		}
	case types.EventTypeTopic:
		if stateKey != "" {
			return nil, errors.New("state key must be empty for state " + eventType)
		}
	case types.EventTypeJoinRules:
		if stateKey != "" {
			return nil, errors.New("state key must be empty for state " + eventType)
		}
	case types.EventTypePowerLevels:
		if stateKey != "" {
			return nil, errors.New("state key must be empty for state " + eventType)
		}
	case types.EventTypeCreate:
		return nil, errors.New("cannot set state " + eventType)

	case types.EventTypeAliases:
		return nil, errors.New("cannot set state " + eventType)

	case types.EventTypeMembership:
		membership, ok := content.(*types.MembershipEventContent)
		if !ok || membership == nil {
			panic("expected membership event content, got " + reflect.TypeOf(content).String())
		}
		if !isUserIdStateKey {
			return nil, errors.New("state key must be a user id for state " + eventType)
		}
		return r.doMembershipChange(user, userIdStateKey, membership)
	}
	if isUserIdStateKey && userIdStateKey != user.Id() {
		return nil, errors.New("cannot set the state of another user")
	}
	return nil, nil
}

func (r Room) doMembershipChange(changeBy User, userId types.UserId, membership *types.MembershipEventContent) (*types.State, error) {
	currentMembership, err := r.UserMembership(userId)
	if err != nil {
		return nil, err
	}
	if currentMembership == membership.Membership {
		return nil, errors.New("membership change was a no-op")
	}
	membership.UserProfile = nil

	switch membership.Membership {
	case types.MembershipNone:
		if currentMembership != types.MembershipBanned {
			return nil, errors.New("invalid or missing membership in membership change")
		}
		err = r.TestPowerLevel(changeBy.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
		if userId == changeBy.Id() {
			return nil, errors.New("cannot remove a ban from self")
		}

	case types.MembershipInvited:
		if currentMembership != types.MembershipNone {
			return nil, errors.New("could not invite user to room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := r.AllowsJoinRule(types.JoinRuleInvite)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("room does not allow join method: " + types.JoinRuleInvite.String())
		}
		err = r.TestPowerLevel(changeBy.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Invite
		})
		if err != nil {
			return nil, err
		}

	case types.MembershipMember:
		if userId != changeBy.Id() {
			return nil, errors.New("cannot force other users to join the room")
		}
		profile, err := changeBy.GetProfile()
		if err != nil {
			return nil, err
		}
		membership.UserProfile = &profile

	case types.MembershipKnocking:
		if userId != changeBy.Id() {
			return nil, errors.New("cannot force other users to knock")
		}
		if currentMembership != types.MembershipNone {
			return nil, errors.New("could not knock on room, already have membership '" + currentMembership.String() + "'")
		}
		ok, err := r.AllowsJoinRule(types.JoinRuleKnock)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("room does not allow join method: " + types.JoinRuleKnock.String())
		}

	case types.MembershipLeaving:
		if currentMembership == types.MembershipNone {
			return nil, errors.New("tried to leave a room without current membership")
		}
		if currentMembership == types.MembershipBanned {
			return nil, errors.New("tried to leave room with current membership '" + types.MembershipBanned.String() + "'")
		}
		if userId != changeBy.Id() {
			err = r.TestPowerLevel(changeBy.Id(), func(pl *types.PowerLevelsEventContent) int {
				return pl.Kick
			})
			if err != nil {
				return nil, err
			}
		}

	case types.MembershipBanned:
		if userId == changeBy.Id() {
			return nil, errors.New("cannot ban self")
		}
		err = r.TestPowerLevel(changeBy.Id(), func(pl *types.PowerLevelsEventContent) int {
			return pl.Ban
		})
		if err != nil {
			return nil, err
		}
	}
	return db.SetRoomState(r.Id(), changeBy.Id(), membership, userId.String())
}

func (r Room) TestPowerLevel(userId types.UserId, powerLevelFunc func(*types.PowerLevelsEventContent) int) error {
	powerLevels, err := r.PowerLevels()
	if err != nil {
		return err
	}
	userPowerLevel, err := r.UserPowerLevel(userId)
	if err != nil {
		return err
	}
	requiredPowerLevel := powerLevelFunc(powerLevels)
	if userPowerLevel < requiredPowerLevel {
		msg := fmt.Sprintf("not enough power level to perform action (%d < %d)", userPowerLevel, requiredPowerLevel)
		return errors.New(msg)
	}
	return nil
}

func (r Room) UserMembership(userId types.UserId) (types.Membership, error) {
	state, err := db.GetRoomState(r.Id(), types.EventTypeMembership, userId.String())
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

func (r Room) AllowsJoinRule(joinRule types.JoinRule) (bool, error) {
	state, err := db.GetRoomState(r.Id(), types.EventTypeJoinRules, "")
	if err != nil {
		return false, err
	}
	if state == nil {
		return false, errors.New("room power levels are invalid or missing")
	}
	joinRules, ok := state.Content.(*types.JoinRulesEventContent)
	if !ok {
		panic("invalid join rule content, was " + reflect.TypeOf(state.Content).String())
	}
	if joinRules.JoinRule != joinRule {
		return false, errors.New("room does not allow join rule: " + joinRule.String())
	}
	return true, nil
}

func (r Room) PowerLevels() (*types.PowerLevelsEventContent, error) {
	state, err := db.GetRoomState(r.Id(), types.EventTypePowerLevels, "")
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, errors.New("room power levels are invalid or missing")
	}
	powerLevels, ok := state.Content.(*types.PowerLevelsEventContent)
	if !ok {
		panic("invalid power level content, was " + reflect.TypeOf(state.Content).String())
	}
	return powerLevels, nil
}

func (r Room) UserPowerLevel(userId types.UserId) (int, error) {
	powerLevels, err := r.PowerLevels()
	if err != nil {
		return 0, err
	}
	if userLevel, ok := powerLevels.Users[userId]; ok {
		return userLevel, nil
	}
	return powerLevels.UserDefault, nil
}

func (r Room) EventPowerLevel(eventType string) (int, error) {
	powerLevels, err := r.PowerLevels()
	if err != nil {
		return 0, err
	}
	if eventLevel, ok := powerLevels.Events[eventType]; ok {
		return eventLevel, nil
	}
	return powerLevels.EventDefault, nil
}
