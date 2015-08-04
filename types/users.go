package types

import (
	"errors"
	"strconv"
	"time"
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

type UserPresence struct {
	Presence      Presence   `json:"presence"`
	StatusMessage string     `json:"status_msg"`
	LastActive    LastActive `json:"last_active_ago"`
}

type User struct {
	UserProfile
	UserPresence
	UserId UserId `json:"user_id"`
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
