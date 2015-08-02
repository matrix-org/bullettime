package types

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Rugvip/bullettime/utils"
)

type IdPrefix rune

type idInterface interface {
	prefix() IdPrefix
	tag() string
	setTag(string)
	hostname() string
	setHostname(string)
}

const (
	UserIdPrefix  IdPrefix = '@'
	RoomIdPrefix           = '!'
	EventIdPrefix          = '$'
	AliasPrefix            = '#'
)

type Id struct {
	Tag      string
	Hostname string
}

func (id *Id) tag() string {
	return id.Tag
}
func (id *Id) setTag(tag string) {
	id.Tag = tag
}
func (id *Id) hostname() string {
	return id.Hostname
}
func (id *Id) setHostname(hostname string) {
	id.Hostname = hostname
}

func parseId(id idInterface, str string) (err error) {
	if len(str) < 2 {
		return errors.New("invalid id, too short, '" + str + "'")
	}
	parsedPrefix, prefixSize := utf8.DecodeRuneInString(str)
	prefix := rune(id.prefix())
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
	id.setTag(split[0])
	id.setHostname(split[1])
	return nil
}

func stringifyId(id idInterface) string {
	return fmt.Sprintf("%c%s:%s", id.prefix(), id.tag(), id.hostname())
}

type UserId struct{ Id }
type RoomId struct{ Id }
type EventId struct{ Id }
type Alias struct{ Id }

func NewRoomId(tag, hostname string) RoomId {
	return RoomId{Id{tag, hostname}}
}

func NewAlias(tag, hostname string) Alias {
	return Alias{Id{tag, hostname}}
}

func NewEventId(tag, hostname string) EventId {
	return EventId{Id{tag, hostname}}
}

func NewUserId(tag, hostname string) UserId {
	return UserId{Id{tag, hostname}}
}

func ParseUserId(str string) (id UserId, err error) {
	err = parseId(&id, str)
	return id, err
}

func ParseRoomId(str string) (id RoomId, err error) {
	err = parseId(&id, str)
	return id, err
}

func ParseEventId(str string) (id EventId, err error) {
	err = parseId(&id, str)
	return id, err
}

func ParseAlias(str string) (id Alias, err error) {
	err = parseId(&id, str)
	return id, err
}

func (id *UserId) prefix() IdPrefix  { return '@' }
func (id *RoomId) prefix() IdPrefix  { return '!' }
func (id *EventId) prefix() IdPrefix { return '$' }
func (id *Alias) prefix() IdPrefix   { return '#' }

func (id *UserId) String() string  { return stringifyId(id) }
func (id *RoomId) String() string  { return stringifyId(id) }
func (id *EventId) String() string { return stringifyId(id) }
func (id *Alias) String() string   { return stringifyId(id) }

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

func (id *UserId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id)), nil
}
func (id *RoomId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id)), nil
}
func (id *EventId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id)), nil
}
func (id *Alias) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id)), nil
}
