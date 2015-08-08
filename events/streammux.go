package events

import (
	"log"
	"sync"

	"github.com/Rugvip/bullettime/types"
)

func NewStreamMux() (*streamMux, types.Error) {
	return &streamMux{
		channels: map[types.UserId]userChannels{},
	}, nil
}

type streamMux struct {
	sync.RWMutex
	channels map[types.UserId]userChannels
}

type userChannels []chan types.IndexedEvent

type indexedEvent struct {
	types.Event
	index uint64
}

func (e indexedEvent) Index() uint64 {
	return e.index
}

func (chs *userChannels) send(event types.IndexedEvent) types.Error {
	for _, ch := range *chs {
		ch <- event
		close(ch)
	}
	*chs = (*chs)[:0]
	return nil
}

func (chs *userChannels) close(ch chan types.IndexedEvent) {
	l := len(*chs)
	for i, channel := range *chs {
		if ch == channel {
			close(channel)
			(*chs)[i] = (*chs)[l-1]
			(*chs)[l-1] = nil
			*chs = (*chs)[:l-1]
			break
		}
	}
}

func (chs *userChannels) make() chan types.IndexedEvent {
	channel := make(chan types.IndexedEvent, 1)
	*chs = append(*chs, channel)
	return channel
}

func (s streamMux) Listen(userId types.UserId, cancel chan struct{}) (chan types.IndexedEvent, types.Error) {
	s.Lock()
	defer s.Unlock()
	chs, ok := s.channels[userId]
	channel := chs.make()
	s.channels[userId] = chs
	if !ok {
	}
	log.Println("listen", userId, chs, s.channels[userId])
	go func() {
		log.Println(userId.Id.Id, "cancel?")
		<-cancel
		log.Println(userId.Id.Id, "cancel!")
		s.Lock()
		defer s.Unlock()
		if chs2, ok := s.channels[userId]; ok {
			log.Println(userId.Id.Id, "cancel!", chs2)
			chs2.close(channel)
			log.Println(userId.Id.Id, "cancel!", chs2)
			if len(chs2) == 0 {
				delete(s.channels, userId)
			} else {
				s.channels[userId] = chs2
			}
		}
	}()
	return channel, nil
}

func (s streamMux) Send(userIds []types.UserId, event types.Event, index uint64) types.Error {
	indexed := indexedEvent{event, index}
	s.Lock()
	defer s.Unlock()
	for _, userId := range userIds {
		chs := s.channels[userId]
		if err := chs.send(indexed); err != nil {
			return err
		}
		delete(s.channels, userId)
	}
	return nil
}
