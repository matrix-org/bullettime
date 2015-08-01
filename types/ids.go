package types

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Rugvip/bullettime/utils"
)

type IdPrefix rune

const (
	RoomIdPrefix  IdPrefix = '!'
	AliasPrefix            = '#'
	EventIdPrefix          = '$'
	UserIdPrefix           = '@'
)

type Id struct {
	prefix   IdPrefix
	Tag      string
	Hostname string
}

func NewRoomId(tag, hostname string) RoomId {
	return RoomId{Id{RoomIdPrefix, tag, hostname}}
}

func NewAlias(tag, hostname string) Alias {
	return Alias{Id{AliasPrefix, tag, hostname}}
}

func NewEventId(tag, hostname string) EventId {
	return EventId{Id{EventIdPrefix, tag, hostname}}
}

func NewUserId(tag, hostname string) UserId {
	return UserId{Id{UserIdPrefix, tag, hostname}}
}

func parseIdWithPrefix(str string, prefix IdPrefix) (id Id, err error) {
	if len(str) < 2 {
		return id, errors.New("invalid id, too short, '" + str + "'")
	}
	parsedPrefix, prefixSize := utf8.DecodeRuneInString(str)
	if parsedPrefix != rune(prefix) {
		msg := fmt.Sprintf("invalid id prefix, was '%c', should be '%c'", parsedPrefix, prefix)
		return id, errors.New(msg)
	}
	rest := str[prefixSize:]
	split := strings.Split(rest, ":")
	if len(split) != 2 {
		msg := fmt.Sprintf("invalid id, should contain exactly one ':', contained %d", len(split)-1)
		return id, errors.New(msg)
	}
	id.prefix = prefix
	id.Tag = split[0]
	id.Hostname = split[1]
	return
}

func ParseUserId(str string) (UserId, error) {
	id, err := parseIdWithPrefix(str, UserIdPrefix)
	if err != nil {
		return UserId{}, err
	}
	return UserId{id}, nil
}

func ParseRoomId(str string) (RoomId, error) {
	id, err := parseIdWithPrefix(str, RoomIdPrefix)
	if err != nil {
		return RoomId{}, err
	}
	return RoomId{id}, nil
}

func ParseEventId(str string) (EventId, error) {
	id, err := parseIdWithPrefix(str, EventIdPrefix)
	if err != nil {
		return EventId{}, err
	}
	return EventId{id}, nil
}

func ParseAlias(str string) (Alias, error) {
	id, err := parseIdWithPrefix(str, AliasPrefix)
	if err != nil {
		return Alias{}, err
	}
	return Alias{id}, nil
}

func (id *Id) String() string {
	return fmt.Sprintf("%c%s:%s", id.prefix, id.Tag, id.Hostname)
}

func (id *Id) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id)), nil
}

type UserId struct{ Id }
type RoomId struct{ Id }
type EventId struct{ Id }
type Alias struct{ Id }

func (id *UserId) UnmarshalJSON(bytes []byte) (err error) {
	*id, err = ParseUserId(utils.StripQuotes(string(bytes)))
	return
}

func (id *RoomId) UnmarshalJSON(bytes []byte) (err error) {
	*id, err = ParseRoomId(utils.StripQuotes(string(bytes)))
	return
}

func (id *EventId) UnmarshalJSON(bytes []byte) (err error) {
	*id, err = ParseEventId(utils.StripQuotes(string(bytes)))
	return
}

func (id *Alias) UnmarshalJSON(bytes []byte) (err error) {
	*id, err = ParseAlias(utils.StripQuotes(string(bytes)))
	return
}
