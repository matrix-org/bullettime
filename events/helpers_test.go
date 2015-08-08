package events

import "github.com/Rugvip/bullettime/types"

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
