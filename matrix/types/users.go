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

package types

import (
	"errors"
	"strconv"
	"time"

	ct "github.com/matrix-org/bullettime/core/types"
)

type UserProfile struct {
	DisplayName string `json:"displayname"`
	AvatarUrl   string `json:"avatar_url"`
}

type Presence int

const (
	PresenceOffline     Presence = 0
	PresenceOnline      Presence = 1
	PresenceAvailable   Presence = 2
	PresenceUnavailable Presence = 3
)

type LastActive time.Time

type UserStatus struct {
	Presence      Presence   `json:"presence"`
	StatusMessage string     `json:"status_msg"`
	LastActive    LastActive `json:"last_active_ago"`
}

type User struct {
	UserProfile
	UserStatus
	UserId ct.UserId `json:"user_id"`
}

func (p *Presence) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "\"offline\"":
		*p = PresenceOffline
		return nil
	case "\"online\"":
		*p = PresenceOnline
		return nil
	case "\"free_for_chat\"":
		*p = PresenceAvailable
		return nil
	case "\"unavailable\"":
		*p = PresenceUnavailable
		return nil
	}
	return errors.New("invalid presence: " + str)
}

func (p Presence) MarshalJSON() ([]byte, error) {
	switch p {
	default:
		return []byte("\"offline\""), nil
	case PresenceOffline:
		return []byte("\"offline\""), nil
	case PresenceOnline:
		return []byte("\"online\""), nil
	case PresenceAvailable:
		return []byte("\"free_for_chat\""), nil
	case PresenceUnavailable:
		return []byte("\"unavailable\""), nil
	}
}

func (l LastActive) MarshalJSON() ([]byte, error) {
	duration := time.Since(time.Time(l)).Nanoseconds() / time.Millisecond.Nanoseconds()
	return []byte(strconv.FormatInt(duration, 10)), nil
}

func (l *LastActive) UnmarshalJSON(data []byte) error {
	return (*time.Time)(l).UnmarshalJSON(data)
}
