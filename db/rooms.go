package db

import (
	"time"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"github.com/Rugvip/bullettime/utils"
)

type roomDb struct {
	events  map[types.EventId]types.Event
	rooms   map[types.RoomId]*dbRoom
	aliases map[types.Alias]*dbRoom
}

func NewRoomDb() (interfaces.RoomStore, types.Error) {
	return &roomDb{
		events:  map[types.EventId]types.Event{},
		rooms:   map[types.RoomId]*dbRoom{},
		aliases: map[types.Alias]*dbRoom{},
	}, nil
}

type stateId struct {
	EventType string
	StateKey  string
}

type dbRoom struct {
	id     types.RoomId
	states map[stateId]*types.State
	events []types.Event
}

func (db *roomDb) CreateRoom(hostname string, alias *types.Alias) (id types.RoomId, err types.Error) {
	if alias != nil && db.aliases[*alias] != nil {
		err = types.RoomInUseError("room alias '" + alias.String() + "' already exists")
		return
	}
	id.Domain = hostname
	for {
		id.Id.Id = utils.RandomString(16)
		if db.rooms[id] == nil {
			break
		}
	}
	db.rooms[id] = &dbRoom{
		id:     id,
		states: map[stateId]*types.State{},
	}
	if alias != nil {
		db.aliases[*alias] = db.rooms[id]
	}
	return
}

func (db *roomDb) RoomExists(id types.RoomId) types.Error {
	if db.rooms[id] == nil {
		return types.NotFoundError("room '" + id.String() + "' doesn't exist")
	}
	return nil
}

func (db *roomDb) AddRoomMessage(roomId types.RoomId, userId types.UserId, content types.TypedContent) (*types.Message, types.Error) {
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	var eventId = types.EventId{types.Id{Domain: userId.Domain}}
	for {
		eventId.Id.Id = utils.RandomString(16)
		if db.events[eventId] == nil {
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

	db.events[eventId] = event
	room.events = append(room.events, event)

	return event, nil
}

func (db *roomDb) SetRoomState(roomId types.RoomId, userId types.UserId, content types.TypedContent, stateKey string) (*types.State, types.Error) {
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	var eventId = types.EventId{types.Id{Domain: userId.Domain}}
	for {
		eventId.Id.Id = utils.RandomString(16)
		if db.events[eventId] == nil {
			break
		}
	}
	stateId := stateId{content.EventType(), stateKey}

	state := new(types.State)
	state.EventId = eventId
	state.RoomId = roomId
	state.UserId = userId
	state.EventType = content.EventType()
	state.StateKey = stateKey
	state.Timestamp = types.Timestamp{time.Now()}
	state.Content = content
	state.OldState = (*types.OldState)(room.states[stateId])

	db.events[eventId] = state
	room.events = append(room.events, state)
	room.states[stateId] = state

	return state, nil
}

func (db *roomDb) RoomState(roomId types.RoomId, eventType, stateKey string) (*types.State, types.Error) {
	room := db.rooms[roomId]
	if room == nil {
		return nil, types.NotFoundError("room '" + roomId.String() + "' doesn't exist")
	}
	state := room.states[stateId{eventType, stateKey}]
	return state, nil
}
