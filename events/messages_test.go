// Copyright 2015 OpenMarket Ltd
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
	"fmt"
	"testing"
	"time"

	"github.com/Rugvip/bullettime/db"

	"github.com/Rugvip/bullettime/types"
)

func TestMessageSource(t *testing.T) {
	members, err := db.NewMembershipDb()
	if err != nil {
		t.Fatal(err)
	}
	streamMux, err := NewStreamMux()
	if err != nil {
		t.Fatal(err)
	}
	_es, err := NewMessageSource(members, streamMux)
	if err != nil {
		t.Fatal(err)
	}
	es := MessageSourceTest{_es, t}
	es.push(message("event1", "user1"), 0)
	es.push(message("event1", "user2"), 1)
	es.push(message("event1", "user3"), 2)
	es.check(0, 3, 5, "user3")
	es.check(0, 3, 3, "user3")
	es.check(0, 3, 1, "user3")
	es.check(0, 3, 0)
	es.check(1, 3, 1, "user3")
	es.check(2, 3, 1, "user3")
	es.check(3, 2, 1, "user3")
	es.check(3, 0, 3, "user3")
	es.check(3, 0, 0)
	es.push(message("event2", "user4"), 3)
	es.push(message("event2", "user5"), 4)
	es.push(message("event2", "user6"), 5)
	es.check(0, 6, 2, "user3", "user6")
	es.check(0, 6, 1, "user3")
	es.check(0, 6, 0)
	es.check(0, 5, 1, "user3")
	es.push(message("event7", "user7"), 6)
	es.check(2, 7, 5, "user3", "user6", "user7")
	es.check(3, 7, 5, "user6", "user7")
}

type MessageSourceTest struct {
	*messageSource
	t *testing.T
}

func (es MessageSourceTest) push(event *types.Message, expectedIndex uint64) {
	index, err := es.Send(event)
	if err != nil {
		es.t.Fatal(err)
	}
	if index != expectedIndex {
		es.t.Fatal("index should be", expectedIndex, "was", index)
	}
}

func (es MessageSourceTest) check(from, to uint64, limit int, expect ...string) {
	user := types.NewUserId("test", "test")
	roomSet := map[types.RoomId]struct{}{
		types.NewRoomId("room", "test"): struct{}{},
	}
	result, err := es.Range(user, roomSet, from, to, limit)
	if err != nil {
		es.t.Fatal(err)
	}
	str := fmt.Sprintf("{from=%v, to=%v, limit=%v, expect=%v}", from, to, limit, expect)
	if len(result) != len(expect) {
		es.t.Fatal(str+": result length should be", len(expect), "was", len(result))
	}
	for i := range result {
		id := result[i].GetContent().(types.CreateEventContent).Creator.Id.Id
		if id != expect[i] {
			es.t.Fatal(str+": result", i, "should be", expect[i], "was", id)
		}
	}
}

func message(eventId, userId string) *types.Message {
	event := types.Message{}
	event.EventType = "m.room.create"
	event.Content = types.CreateEventContent{types.NewUserId(userId, "test")}
	event.RoomId = types.NewRoomId("room", "test")
	event.Timestamp = types.Timestamp{time.Now()}
	event.EventId = types.NewEventId(eventId, "test")
	return &event
}
