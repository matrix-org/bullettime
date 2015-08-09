package db

import (
	"fmt"
	"sync"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

type memberDb struct { // always lock in the same order as below
	sync.RWMutex
	users   map[types.RoomId][]types.UserId
	rooms   map[types.UserId][]types.RoomId
	members map[memberKey]struct{}
}

type memberKey struct {
	types.RoomId
	types.UserId
}

func NewMembershipDb() (interfaces.MembershipStore, error) {
	return &memberDb{
		users:   map[types.RoomId][]types.UserId{},
		rooms:   map[types.UserId][]types.RoomId{},
		members: map[memberKey]struct{}{},
	}, nil
}

func (db *memberDb) AddMember(roomId types.RoomId, userId types.UserId) types.Error {
	db.Lock()
	defer db.Unlock()
	member := memberKey{roomId, userId}
	if _, ok := db.members[member]; ok {
		msg := fmt.Sprintf("user %s is already a member of the room %s", userId, roomId)
		return types.ServerError(msg)
	}
	db.members[member] = struct{}{}
	db.users[roomId] = append(db.users[roomId], userId)
	db.rooms[userId] = append(db.rooms[userId], roomId)
	return nil
}

func (db *memberDb) RemoveMember(roomId types.RoomId, userId types.UserId) types.Error {
	db.Lock()
	defer db.Unlock()
	member := memberKey{roomId, userId}
	if _, ok := db.members[member]; !ok {
		msg := fmt.Sprintf("user %s is not a member of the room %s", userId, roomId)
		return types.ServerError(msg)
	}
	users := db.users[roomId]
	for i, l := 0, len(users); i < l; i += 1 {
		if users[i] == userId {
			users[i] = users[l-1]
			users[l-1] = types.UserId{}
			users = users[:l-1]
			break
		}
	}
	rooms := db.rooms[userId]
	for i, l := 0, len(rooms); i < l; i += 1 {
		if rooms[i] == roomId {
			rooms[i] = rooms[l-1]
			rooms[l-1] = types.RoomId{}
			rooms = rooms[:l-1]
			break
		}
	}
	delete(db.members, member)
	return nil
}

func (db *memberDb) Rooms(userId types.UserId) ([]types.RoomId, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.rooms[userId], nil
}

func (db *memberDb) Users(roomId types.RoomId) ([]types.UserId, types.Error) {
	db.RLock()
	defer db.RUnlock()
	return db.users[roomId], nil
}

func (db *memberDb) Peers(user types.UserId) ([]types.UserId, types.Error) {
	db.RLock()
	defer db.RUnlock()
	peerSet := map[types.UserId]struct{}{}
	for _, room := range db.rooms[user] {
		for _, peer := range db.users[room] {
			peerSet[peer] = struct{}{}
		}
	}
	peers := make([]types.UserId, len(peerSet))
	for peer := range peerSet {
		peers = append(peers, peer)
	}
	return peers, nil
}
