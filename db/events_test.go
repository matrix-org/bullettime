package db

import (
	"testing"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

func TestEventStream(t *testing.T) {
	_es, err := NewEventStream()
	if err != nil {
		t.Fatal(err)
	}
	es := EventStreamTest{_es, t}
	es.push(typing("room1", "user1"), 1)
	es.push(typing("room1", "user2"), 2)
	es.push(typing("room1", "user3"), 3)
	es.check(0, "user3")
	es.check(2, "user3")
	es.check(3)
	es.check(4)
	es.push(typing("room2", "user4"), 4)
	es.push(typing("room2", "user5"), 5)
	es.push(typing("room2", "user6"), 6)
	es.check(0, "user6", "user3")
	es.check(3, "user6")
	es.push(typing("room1", "user7"), 7)
	es.check(3, "user7", "user6")
	es.check(6, "user7")
	es.check(7)
	es.check(8)
}

type EventStreamTest struct {
	interfaces.EventStream
	t *testing.T
}

func (es EventStreamTest) push(event types.Event, expectedIndex uint64) {
	index, err := es.Push(event)
	if err != nil {
		es.t.Fatal(err)
	}
	if index != expectedIndex {
		es.t.Error("index should be", expectedIndex, "was", index)
	}
}

func (es EventStreamTest) check(from uint64, expect ...string) {
	result, err := es.Iterate(from)
	if err != nil {
		es.t.Fatal(err)
	}
	if len(result) != len(expect) {
		es.t.Error("result length should be", len(expect), "was", len(result))
	}
	for i := range result {
		id := result[i].GetContent().(types.TypingEventContent).UserIds[0].Id.Id
		if id != expect[i] {
			es.t.Error("result", i, "should be", expect[i], "was", id)
		}
	}
}

func typing(id string, ids ...string) types.Event {
	event := &types.TypingEvent{}
	event.EventType = "m.typing"
	event.RoomId = types.NewRoomId(id, "test")
	userIds := make([]types.UserId, len(ids))
	for i := range ids {
		userIds[i] = types.NewUserId(ids[i], "test")
	}
	event.Content.UserIds = userIds
	return event
}
