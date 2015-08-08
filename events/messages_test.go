package events

import (
	"testing"
	"time"

	"github.com/Rugvip/bullettime/types"
)

func TestMessageSource(t *testing.T) {
	_es, err := NewMessageSource()
	if err != nil {
		t.Fatal(err)
	}
	es := MessageSourceTest{_es, t}
	es.push(message("event1", "user1"), 0)
	es.push(message("event1", "user2"), 1)
	es.push(message("event1", "user3"), 2)
	es.check(0, "user3")
	es.check(1, "user3")
	es.check(2)
	es.check(3)
	es.push(message("event2", "user4"), 3)
	es.push(message("event2", "user5"), 4)
	es.push(message("event2", "user6"), 5)
	es.check(0, "user6", "user3")
	es.check(2, "user6")
	es.push(message("event7", "user7"), 6)
	es.check(2, "user7", "user6")
	es.check(5, "user7")
	es.check(6)
	es.check(7)
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

func (es MessageSourceTest) check(from uint64, expect ...string) {
	result, err := es.Iterate(from)
	if err != nil {
		es.t.Fatal(err)
	}
	if len(result) != len(expect) {
		es.t.Fatal("result length should be", len(expect), "was", len(result))
	}
	for i := range result {
		id := result[i].GetContent().(types.CreateEventContent).Creator.Id.Id
		if id != expect[i] {
			es.t.Fatal("result", i, "should be", expect[i], "was", id)
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
