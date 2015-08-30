package service

import (
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

func NewPresenceService(
	presenceProvider interfaces.PresenceProvider,
	presenceEventSink interfaces.PresenceEventSink,
) (interfaces.PresenceService, error) {
	return presenceService{
		presenceProvider,
		presenceEventSink,
	}, nil
}

type presenceService struct {
	presenceProvider  interfaces.PresenceProvider
	presenceEventSink interfaces.PresenceEventSink
}

func (s presenceService) Status(user, caller types.UserId) (types.UserStatus, types.Error) {
	return s.presenceProvider.Status(user)
}

func (s presenceService) UpdateStatus(
	user, caller types.UserId,
	presence *types.Presence,
	statusMessage *string,
) (types.UserStatus, types.Error) {
	if user != caller {
		return types.UserStatus{}, types.ForbiddenError("can't change the presence of other users")
	}
	status, err := s.presenceProvider.Status(user)
	if err != nil {
		return types.UserStatus{}, err
	}
	if presence != nil {
		status.Presence = *presence
	}
	if statusMessage != nil {
		status.StatusMessage = *statusMessage
	}
	_, err = s.presenceEventSink.SetUserStatus(user, status)
	if err != nil {
		return types.UserStatus{}, err
	}
	return status, nil
}
