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
	"log"

	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	"github.com/matrix-org/bullettime/matrix/types"
)

func NewProfileService(
	profiles interfaces.ProfileProvider,
	profileSink interfaces.ProfileEventSink,
	members interfaces.MembershipStore,
	rooms interfaces.RoomStore,
	eventSink interfaces.EventSink,
) (interfaces.ProfileService, error) {
	return profileService{
		profiles,
		profileSink,
		members,
		rooms,
		eventSink,
	}, nil
}

type profileService struct {
	profiles    interfaces.ProfileProvider
	profileSink interfaces.ProfileEventSink
	members     interfaces.MembershipStore
	rooms       interfaces.RoomStore
	eventSink   interfaces.EventSink
}

func (s profileService) Profile(user, caller ct.UserId) (types.UserProfile, ct.Error) {
	return s.profiles.Profile(user)
}

func (s profileService) UpdateProfile(
	user, caller ct.UserId,
	name, avatarUrl *string,
) (types.UserProfile, ct.Error) {
	if user != caller {
		return types.UserProfile{}, types.ForbiddenError("can't change the profile of other users")
	}
	profile, err := s.profiles.Profile(user)
	if err != nil {
		return types.UserProfile{}, err
	}
	if name != nil {
		profile.DisplayName = *name
	}
	if avatarUrl != nil {
		profile.AvatarUrl = *avatarUrl
	}
	_, err = s.profileSink.SetUserProfile(user, profile)
	if err != nil {
		return types.UserProfile{}, err
	}
	rooms, err := s.members.Rooms(user)
	if err != nil {
		return types.UserProfile{}, err
	}
	log.Printf("GOT LE STUFF %s, %#v", user, rooms)
	for _, room := range rooms {
		membership, err := s.rooms.RoomState(room, types.EventTypeMembership, user.String())
		if err != nil {
			log.Println("failed to get membership when updating profile: " + err.Error())
			continue
		}
		content := *membership.Content.(*types.MembershipEventContent)
		content.UserProfile = &profile
		state, err := s.rooms.SetRoomState(room, user, &content, user.String()) //TODO: fix race, CAS?
		if err != nil {
			log.Println("failed to set membership when updating profile: " + err.Error())
			continue
		}
		_, err = s.eventSink.Send(state)
		if err != nil {
			log.Println("failed to send event when updating profile: " + err.Error())
		}
	}
	return profile, nil
}
