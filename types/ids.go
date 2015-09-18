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
	"fmt"
	"log"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/matrix-org/bullettime/utils"
)

const (
	UserIdPrefix  = '@'
	RoomIdPrefix  = '!'
	EventIdPrefix = '$'
	AliasPrefix   = '#'
)

type Id struct {
	Prefix rune
	Id     string
	domain int
}

type IdParseError string

func (e IdParseError) Error() string {
	return "failed to parse id: " + string(e)
}

func parseId(prefix rune, id *Id, str string) error {
	if len(str) < 2 {
		return IdParseError("too short")
	}
	parsedPrefix, prefixSize := utf8.DecodeRuneInString(str)
	if parsedPrefix != prefix {
		msg := fmt.Sprintf("prefix was '%c', should have been '%c'", parsedPrefix, prefix)
		return IdParseError(msg)
	}
	rest := str[prefixSize:]
	split := strings.Split(rest, ":")
	if len(split) != 2 {
		msg := fmt.Sprintf("should contain exactly one ':', contained %d", len(split)-1)
		return IdParseError(msg)
	}
	if split[0] == "" {
		return IdParseError("missing id part")
	}
	if split[1] == "" {
		return IdParseError("missing domain part")
	}
	id.Prefix = prefix
	id.Id = split[0]
	id.domain = domainId(split[1])
	return nil
}

func (id Id) String() string {
	if !id.Valid() {
		panic("tried to stringify invalid id: {" + id.Id + ", " + id.Domain() + "}")
	}
	return fmt.Sprintf("%c%s:%s", id.Prefix, id.Id, id.Domain())
}

func (id Id) Valid() bool {
	return id.Prefix != 0 && id.Id != "" && id.Domain() != ""
}

func (id Id) Domain() string {
	return domainName(id.domain)
}

type UserId Id
type RoomId Id
type EventId Id
type Alias Id

func NewRoomId(id, domain string) RoomId {
	return RoomId{RoomIdPrefix, id, domainId(domain)}
}

func NewAlias(id, domain string) Alias {
	return Alias{AliasPrefix, id, domainId(domain)}
}

func NewEventId(id, domain string) EventId {
	return EventId{EventIdPrefix, id, domainId(domain)}
}

func NewUserId(id, domain string) UserId {
	return UserId{UserIdPrefix, id, domainId(domain)}
}

func DeriveId(id string, from Id) Id {
	return Id{from.Prefix, id, from.domain}
}

func DeriveRoomId(id string, from Id) RoomId {
	return RoomId(Id{RoomIdPrefix, id, from.domain})
}

func DeriveAlias(id string, from Id) Alias {
	return Alias(Id{AliasPrefix, id, from.domain})
}

func DeriveEventId(id string, from Id) EventId {
	return EventId(Id{EventIdPrefix, id, from.domain})
}

func DeriveUserId(id string, from Id) UserId {
	return UserId(Id{UserIdPrefix, id, from.domain})
}

func ParseUserId(str string) (id UserId, err error) {
	err = parseId(UserIdPrefix, (*Id)(&id), str)
	return id, err
}

func ParseRoomId(str string) (id RoomId, err error) {
	err = parseId(RoomIdPrefix, (*Id)(&id), str)
	return id, err
}

func ParseEventId(str string) (id EventId, err error) {
	err = parseId(EventIdPrefix, (*Id)(&id), str)
	return id, err
}

func ParseAlias(str string) (id Alias, err error) {
	err = parseId(AliasPrefix, (*Id)(&id), str)
	return id, err
}

func (i UserId) String() string  { return Id(i).String() }
func (i RoomId) String() string  { return Id(i).String() }
func (i EventId) String() string { return Id(i).String() }
func (i Alias) String() string   { return Id(i).String() }

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
	return []byte(fmt.Sprintf(`"%s"`, i)), nil
}
func (i RoomId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, i)), nil
}
func (i EventId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, i)), nil
}
func (i Alias) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, i)), nil
}

var domainTableLock sync.RWMutex
var domainIdTable = map[string]int{}
var domainNames = []string{""}

func domainId(domain string) int {
	domainTableLock.RLock()
	id, ok := domainIdTable[domain]
	domainTableLock.RUnlock()
	if ok {
		return id
	}
	domainTableLock.Lock()
	defer domainTableLock.Unlock()
	if id, ok := domainIdTable[domain]; ok { // since we had to reacquire the lock
		return id
	}
	id = len(domainNames)
	domainIdTable[domain] = id
	domainNames = append(domainNames, domain)
	return id
}

func domainName(id int) string {
	domainTableLock.RLock()
	defer domainTableLock.RUnlock()
	if id <= 0 || id >= len(domainNames) {
		log.Panicf("invalid domain index: %d, should be [1, %d]", id, len(domainNames))
	}
	return domainNames[id]
}
