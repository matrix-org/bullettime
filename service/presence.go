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
	"github.com/matrix-org/bullettime/interfaces"
	"github.com/matrix-org/bullettime/types"
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
