package db

import (
	"sync"
	"time"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

type aliasDb struct { // always lock in the same order as below
	aliasesLock sync.RWMutex
	aliases     map[types.RoomId][]types.Alias

	roomsLock sync.RWMutex
	rooms     map[types.Alias]types.RoomId
	reserved  map[types.Alias]struct{}
}

func NewAliasDb() (interfaces.AliasStore, types.Error) {
	return &aliasDb{
		aliases:  map[types.RoomId][]types.Alias{},
		rooms:    map[types.Alias]types.RoomId{},
		reserved: map[types.Alias]struct{}{},
	}, nil
}

func (db *aliasDb) Reserve(alias types.Alias) types.Error {
	db.roomsLock.Lock()
	defer db.roomsLock.Unlock()
	if _, ok := db.rooms[alias]; ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' already exists")
	}
	if _, ok := db.reserved[alias]; ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' already reserved")
	}
	db.reserved[alias] = struct{}{}
	go func() {
		time.Sleep(time.Second * 10)
		delete(db.reserved, alias)
	}()
	return nil
}

func (db *aliasDb) Claim(alias types.Alias, roomId types.RoomId) types.Error {
	db.roomsLock.Lock()
	defer db.roomsLock.Unlock()
	if _, ok := db.rooms[alias]; ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' already exists")
	}
	if _, ok := db.reserved[alias]; !ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' was not reserved")
	}
	delete(db.reserved, alias)
	db.rooms[alias] = roomId

	db.aliasesLock.Lock()
	defer db.aliasesLock.Unlock()
	db.aliases[roomId] = append(db.aliases[roomId], alias)
	return nil
}

func (db *aliasDb) AddAlias(alias types.Alias, roomId types.RoomId) types.Error {
	db.roomsLock.Lock()
	defer db.roomsLock.Unlock()
	if _, ok := db.rooms[alias]; ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' already exists")
	}
	if _, ok := db.reserved[alias]; ok {
		return types.RoomInUseError("room alias '" + alias.String() + "' is reserved")
	}
	db.rooms[alias] = roomId

	db.aliasesLock.Lock()
	defer db.aliasesLock.Unlock()
	db.aliases[roomId] = append(db.aliases[roomId], alias)

	return nil
}

func (db *aliasDb) RemoveAlias(alias types.Alias, roomId types.RoomId) types.Error {
	db.roomsLock.Lock()
	defer db.roomsLock.Unlock()
	if _, ok := db.rooms[alias]; !ok {
		return types.NotFoundError("room alias '" + alias.String() + "' doesn't exist")
	}
	delete(db.rooms, alias)

	db.aliasesLock.Lock()
	defer db.aliasesLock.Unlock()

	aliases := db.aliases[roomId]
	l := len(aliases)
	for i := 0; i < l; i += 1 {
		if aliases[i] == alias {
			aliases[i] = aliases[l-1]
			aliases[l-1] = types.Alias{}
			aliases = aliases[:l-1]
			break
		}
	}
	db.aliases[roomId] = aliases
	return nil
}

func (db *aliasDb) Aliases(roomId types.RoomId) ([]types.Alias, types.Error) {
	db.aliasesLock.RLock()
	defer db.aliasesLock.RUnlock()
	return db.aliases[roomId], nil
}

func (db *aliasDb) Room(alias types.Alias) (*types.RoomId, types.Error) {
	db.roomsLock.RLock()
	defer db.roomsLock.RUnlock()
	if roomId, ok := db.rooms[alias]; ok {
		return &roomId, nil
	}
	return nil, nil
}
