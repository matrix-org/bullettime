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

package events

import (
	"testing"

	"github.com/matrix-org/bullettime/core/db"
	"github.com/matrix-org/bullettime/core/events"
	ct "github.com/matrix-org/bullettime/core/types"
	"github.com/matrix-org/bullettime/matrix/interfaces"
	"github.com/matrix-org/bullettime/matrix/service"
	"github.com/matrix-org/bullettime/matrix/stores"
	"github.com/matrix-org/bullettime/matrix/types"
)

type services struct {
	room     interfaces.RoomService
	user     interfaces.UserService
	profile  interfaces.ProfileService
	presence interfaces.PresenceService
	token    interfaces.TokenService
	event    interfaces.EventService
	sync     interfaces.SyncService
}

func setup() services {
	stateStore, err := db.NewStateStore()
	if err != nil {
		panic(err)
	}
	roomStore, err := db.NewRoomDb()
	if err != nil {
		panic(err)
	}
	userStore, err := stores.NewUserDb(stateStore)
	if err != nil {
		panic(err)
	}
	aliasCache, err := db.NewIdMap()
	if err != nil {
		panic(err)
	}
	aliasStore, err := stores.NewAliasStore(aliasCache)
	if err != nil {
		panic(err)
	}
	memberCache, err := db.NewIdMultiMap()
	if err != nil {
		panic(err)
	}
	memberStore, err := stores.NewMembershipStore(memberCache)
	if err != nil {
		panic(err)
	}
	streamMux, err := events.NewStreamMux()
	if err != nil {
		panic(err)
	}
	messageStream, err := events.NewMessageStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}
	presenceStream, err := events.NewPresenceStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}
	typingStream, err := events.NewTypingStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}

	roomService, err := service.CreateRoomService(
		roomStore,
		aliasStore,
		memberStore,
		messageStream,
		presenceStream,
		typingStream,
		typingStream,
	)
	if err != nil {
		panic(err)
	}
	userService, err := service.CreateUserService(userStore)
	if err != nil {
		panic(err)
	}
	profileService, err := service.NewProfileService(
		presenceStream,
		presenceStream,
		memberStore,
		roomStore,
		messageStream,
	)
	if err != nil {
		panic(err)
	}
	presenceService, err := service.NewPresenceService(presenceStream, presenceStream)
	if err != nil {
		panic(err)
	}
	tokenService, err := service.CreateTokenService()
	if err != nil {
		panic(err)
	}
	eventService, err := service.NewEventService(
		messageStream,
		presenceStream,
		typingStream,
		streamMux,
		messageStream,
		memberStore,
	)
	if err != nil {
		panic(err)
	}
	syncService, err := service.NewSyncService(
		messageStream,
		presenceStream,
		typingStream,
		roomStore,
		memberStore,
	)
	if err != nil {
		panic(err)
	}
	return services{
		roomService,
		userService,
		profileService,
		presenceService,
		tokenService,
		eventService,
		syncService,
	}
}

func TestUserCreation(t *testing.T) {
	s := setup()
	userId := ct.NewUserId("test", "matrix.org")
	if err := s.user.CreateUser(userId); err != nil {
		t.Fatal(err)
	}
	s.user.UserExists(userId, userId)
	err := s.user.CreateUser(userId)
	if err == nil {
		t.Fatal("expected M_USER_IN_USE error")
	} else if err.Code() != "M_USER_IN_USE" {
		t.Error("expected M_USER_IN_USE error code but got ", err.Code())
	}
	status, err := s.presence.Status(userId, userId)
	if err != nil {
		t.Fatal(err)
	}
	if status.Presence != types.PresenceOffline {
		t.Error("expected offline presence")
	}
	if status.StatusMessage != "" {
		t.Error("expected empty status message")
	}
}
