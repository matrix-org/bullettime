package db

import (
	"time"

	"github.com/Rugvip/bullettime/types"
	"github.com/Rugvip/bullettime/utils"
)

type StateId struct {
	EventType string
	StateKey  string
}

type room struct {
	id     types.RoomId
	states map[StateId]*types.State
	events []types.Event
}

var eventTable = map[types.EventId]*types.Event{}

var roomTable = map[types.RoomId]*room{}

var aliasTable = map[types.Alias]*room{}

func CreateRoom(hostname string, alias *types.Alias) (id types.RoomId, err error) {
	if alias != nil && aliasTable[*alias] != nil {
		err = types.RoomInUseError("room alias '" + alias.String() + "' already exists")
		return
	}
	id.Domain = hostname
	for {
		id.Id.Id = utils.RandomString(16)
		if roomTable[id] == nil {
			break
		}
	}
	roomTable[id] = &room{
		id:     id,
		states: map[StateId]*types.State{},
	}
	if alias != nil {
		aliasTable[*alias] = roomTable[id]
	}
	return
}

func RoomExists(id types.RoomId) error {
	if roomTable[id] == nil {
		return types.NotFoundError("room '" + id.String() + "' doesn't exist")
	}
	return nil
}

func AddRoomMessage(roomId types.RoomId, userId types.UserId, content types.TypedContent) (*types.Message, error) {
	room := roomTable[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	var eventId = types.EventId{types.Id{Domain: userId.Domain}}
	for {
		eventId.Id.Id = utils.RandomString(16)
		if eventTable[eventId] == nil {
			break
		}
	}
	event := new(types.Message)
	event.EventId = eventId
	event.RoomId = roomId
	event.UserId = userId
	event.EventType = content.EventType()
	event.Timestamp = types.Timestamp{time.Now()}
	event.Content = content

	room.events = append(room.events, event)

	return event, nil
}

func SetRoomState(roomId types.RoomId, userId types.UserId, content types.TypedContent, stateKey string) (*types.State, error) {
	room := roomTable[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	var eventId = types.EventId{types.Id{Domain: userId.Domain}}
	for {
		eventId.Id.Id = utils.RandomString(16)
		if eventTable[eventId] == nil {
			break
		}
	}
	stateId := StateId{content.EventType(), stateKey}

	state := new(types.State)
	state.EventId = eventId
	state.RoomId = roomId
	state.UserId = userId
	state.EventType = content.EventType()
	state.StateKey = stateKey
	state.Timestamp = types.Timestamp{time.Now()}
	state.Content = content
	state.OldState = (*types.OldState)(room.states[stateId])

	room.events = append(room.events, state)
	room.states[stateId] = state

	return state, nil
}

func GetRoomState(roomId types.RoomId, eventType, stateKey string) (*types.State, error) {
	room := roomTable[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	state := room.states[StateId{eventType, stateKey}]
	return state, nil
}
