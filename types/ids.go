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
	id() string
	setId(string)
	domain() string
	setDomain(string)
}

const (
	UserIdPrefix  IdPrefix = '@'
	RoomIdPrefix  IdPrefix = '!'
	EventIdPrefix IdPrefix = '$'
	AliasPrefix   IdPrefix = '#'
)

type Id struct {
	Id     string
	Domain string
}

func (i *Id) id() string {
	return i.Id
}
func (i *Id) setId(id string) {
	i.Id = id
}
func (i *Id) domain() string {
	return i.Domain
}
func (i *Id) setDomain(domain string) {
	i.Domain = domain
}

func parseId(i idInterface, str string) (err error) {
	if len(str) < 2 {
		return errors.New("invalid id, too short, '" + str + "'")
	}
	parsedPrefix, prefixSize := utf8.DecodeRuneInString(str)
	prefix := rune(i.prefix())
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
	i.setId(split[0])
	i.setDomain(split[1])
	return nil
}

func stringifyId(id idInterface) string {
	return fmt.Sprintf("%c%s:%s", id.prefix(), id.id(), id.domain())
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

func (i *UserId) prefix() IdPrefix  { return '@' }
func (i *RoomId) prefix() IdPrefix  { return '!' }
func (i *EventId) prefix() IdPrefix { return '$' }
func (i *Alias) prefix() IdPrefix   { return '#' }

func (i *UserId) String() string  { return stringifyId(i) }
func (i *RoomId) String() string  { return stringifyId(i) }
func (i *EventId) String() string { return stringifyId(i) }
func (i *Alias) String() string   { return stringifyId(i) }

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

func (i *UserId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i *RoomId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i *EventId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
func (i *Alias) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", i)), nil
}
