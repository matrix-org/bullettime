package types

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Rugvip/bullettime/utils"
)

const (
	UserIdPrefix  = '@'
	RoomIdPrefix  = '!'
	EventIdPrefix = '$'
	AliasPrefix   = '#'
)

type Id struct {
	Id     string
	Domain string
}

func parseId(id *Id, str string, prefix rune) error {
	if len(str) < 2 {
		return errors.New("invalid id, too short, '" + str + "'")
	}
	parsedPrefix, prefixSize := utf8.DecodeRuneInString(str)
	if parsedPrefix != prefix {
		msg := fmt.Sprintf("invalid id prefix, was '%c', should be '%c'", parsedPrefix, prefix)
		return errors.New(msg)
	}
	rest := str[prefixSize:]
	split := strings.Split(rest, ":")
	if len(split) != 2 {
		msg := fmt.Sprintf("invalid id, should contain exactly one ':', contained %d", len(split)-1)
		return errors.New(msg)
	}
	if split[0] == "" {
		return errors.New("invalid id: missing id part: " + str)
	}
	if split[1] == "" {
		return errors.New("invalid id: missing domain part: " + str)
	}
	id.Id = split[0]
	id.Domain = split[1]
	return nil
}

func stringifyId(id Id, prefix rune) string {
	return fmt.Sprintf("%c%s:%s", prefix, id.Id, id.Domain)
}

type UserId struct{ Id }
type RoomId struct{ Id }
type EventId struct{ Id }
type Alias struct{ Id }

func NewRoomId(id, domain string) RoomId {
	return RoomId{Id{id, domain}}
}

func NewAlias(id, domain string) Alias {
	return Alias{Id{id, domain}}
}

func NewEventId(id, domain string) EventId {
	return EventId{Id{id, domain}}
}

func NewUserId(id, domain string) UserId {
	return UserId{Id{id, domain}}
}

func ParseUserId(str string) (id UserId, err error) {
	err = parseId(&id.Id, str, UserIdPrefix)
	return id, err
}

func ParseRoomId(str string) (id RoomId, err error) {
	err = parseId(&id.Id, str, RoomIdPrefix)
	return id, err
}

func ParseEventId(str string) (id EventId, err error) {
	err = parseId(&id.Id, str, EventIdPrefix)
	return id, err
}

func ParseAlias(str string) (id Alias, err error) {
	err = parseId(&id.Id, str, AliasPrefix)
	return id, err
}

func (i UserId) String() string  { return stringifyId(i.Id, UserIdPrefix) }
func (i RoomId) String() string  { return stringifyId(i.Id, RoomIdPrefix) }
func (i EventId) String() string { return stringifyId(i.Id, EventIdPrefix) }
func (i Alias) String() string   { return stringifyId(i.Id, AliasPrefix) }

func (i *UserId) UnmarshalJSON(bytes []byte) (err error) {
	*i, err = ParseUserId(utils.StripQuotes(string(bytes)))
	return
}

func (i *RoomId) UnmarshalJSON(bytes []byte) (err error) {
	*i, err = ParseRoomId(utils.StripQuotes(string(bytes)))
	return
}

func (i *EventId) UnmarshalJSON(bytes []byte) (err error) {
	*i, err = ParseEventId(utils.StripQuotes(string(bytes)))
	return
}

func (i *Alias) UnmarshalJSON(bytes []byte) (err error) {
	*i, err = ParseAlias(utils.StripQuotes(string(bytes)))
	return
}

func (i UserId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i RoomId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i EventId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i Alias) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
