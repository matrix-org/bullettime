package service

import (
	"log"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
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

func (s profileService) Profile(user, caller types.UserId) (types.UserProfile, types.Error) {
	return s.profiles.Profile(user)
}

func (s profileService) UpdateProfile(
	user, caller types.UserId,
	name, avatarUrl *string,
) (types.UserProfile, types.Error) {
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
