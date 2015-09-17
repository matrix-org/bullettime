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

package service

import (
	"log"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

func NewEventService(
	messageSource interfaces.IndexedEventSource,
	presenceSource interfaces.IndexedEventSource,
	typingSource interfaces.IndexedEventSource,
	asyncEventSource interfaces.AsyncEventSource,
	eventProvider interfaces.EventProvider,
	membershipStore interfaces.MembershipStore,
) (interfaces.EventService, error) {
	return &eventService{
		messageSource,
		presenceSource,
		typingSource,
		asyncEventSource,
		eventProvider,
		membershipStore,
	}, nil
}

type eventService struct {
	messageSource    interfaces.IndexedEventSource
	presenceSource   interfaces.IndexedEventSource
	typingSource     interfaces.IndexedEventSource
	asyncEventSource interfaces.AsyncEventSource
	eventProvider    interfaces.EventProvider
	membershipStore  interfaces.MembershipStore
}

func (s eventService) Event(user types.UserId, eventId types.EventId) (types.Event, types.Error) {
	event, err := s.eventProvider.Event(user, eventId)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, types.NotFoundError("event not found: " + event.GetEventKey().String())
	}
	return event, nil
}

func (s eventService) Range(
	user types.UserId,
	from, to *types.StreamToken,
	limit uint,
	cancel chan struct{},
) (chunk *types.EventStreamRange, err types.Error) {
	var eventCh chan types.IndexedEvent

	if from == nil || to == nil || from.MessageIndex > to.MessageIndex {
		eventCh, err = s.asyncEventSource.Listen(user, cancel)
		if err != nil {
			return nil, err
		}
	}

	maxMessage := s.messageSource.Max()
	maxPresence := s.presenceSource.Max()
	maxTyping := s.typingSource.Max()

	var fromMessage uint64
	var fromPresence uint64
	var fromTyping uint64

	if from != nil {
		fromMessage = from.MessageIndex
		if fromMessage > maxMessage {
			fromMessage = maxMessage
		}
		fromPresence = from.PresenceIndex
		if fromPresence > maxPresence {
			fromPresence = maxPresence
		}
		fromTyping = from.TypingIndex
		if fromTyping > maxTyping {
			fromTyping = maxTyping
		}
	} else {
		fromMessage = maxMessage
		fromPresence = maxPresence
		fromTyping = maxTyping
	}

	var toMessage uint64
	var toPresence uint64
	var toTyping uint64

	if to != nil {
		toMessage = to.MessageIndex
		toPresence = to.PresenceIndex
		toTyping = to.TypingIndex
	} else {
		toMessage = maxMessage
		toPresence = maxPresence
		toTyping = maxTyping
	}

	userSet, err := s.membershipStore.Peers(user)
	if err != nil {
		return nil, err
	}

	roomSet := map[types.RoomId]struct{}{}
	rooms, err := s.membershipStore.Rooms(user)
	if err != nil {
		return nil, err
	}
	for _, room := range rooms {
		roomSet[room] = struct{}{}
	}

	messages, err := s.messageSource.Range(&user, userSet, roomSet, fromMessage, toMessage, limit)
	if err != nil {
		return nil, err
	}
	presences, err := s.presenceSource.Range(&user, userSet, roomSet, fromPresence, toPresence, limit)
	if err != nil {
		return nil, err
	}
	typings, err := s.typingSource.Range(&user, userSet, roomSet, fromTyping, toTyping, limit)
	if err != nil {
		return nil, err
	}

	log.Printf("getting events from %d to %d, max %d, %#v", fromMessage, toMessage, maxMessage, eventCh)

	if eventCh != nil {
		blocking := true
		if to != nil && toMessage <= maxMessage && toPresence <= maxPresence && toTyping <= maxTyping {
			blocking = false
		}

		gotEvent := false
		var event types.IndexedEvent
		if blocking && len(messages)+len(presences)+len(typings) == 0 {
			event, gotEvent = <-eventCh
		} else {
			select {
			case event, gotEvent = <-eventCh:
			default:
			}
		}
		log.Printf("async event: %#v blocking: %#v len: %#v", event, blocking, len(messages)+len(presences)+len(typings))

		if gotEvent && uint(len(messages)) < limit {
			eventType := event.Event().GetEventType()
			if eventType == types.EventTypePresence {
				if len(presences) == 0 || presences[len(presences)-1].Index() < event.Index() {
					if to == nil || event.Index() < toPresence {
						presences = append(presences, event)
					}
				}
			} else if eventType == types.EventTypeTyping {
				if len(typings) == 0 || typings[len(typings)-1].Index() < event.Index() {
					if to == nil || event.Index() < toTyping {
						typings = append(typings, event)
					}
				}
			} else {
				if len(messages) == 0 || messages[len(messages)-1].Index() < event.Index() {
					if to == nil || event.Index() < toMessage {
						messages = append(messages, event)
					}
				}
			}
		}
	}

	messageIndex := fromMessage
	presenceIndex := fromPresence
	typingIndex := fromTyping

	if len(messages) > 0 {
		messageIndex = messages[len(messages)-1].Index() + 1
	}
	if len(presences) > 0 {
		presenceIndex = presences[len(presences)-1].Index() + 1
	}
	if len(typings) > 0 {
		typingIndex = typings[len(typings)-1].Index() + 1
	}

	start := types.NewStreamToken(fromMessage, fromPresence, fromTyping)
	end := types.NewStreamToken(messageIndex, presenceIndex, typingIndex)

	events := make([]types.Event, len(messages)+len(presences)+len(typings))

	for i, _ := range events {
		if i < len(messages) {
			events[i] = messages[i].Event()
		} else {
			i -= len(messages)
			if i < len(presences) {
				events[len(messages)+i] = presences[i].Event()
			} else {
				i -= len(presences)
				events[len(messages)+len(presences)+i] = typings[i].Event()
			}
		}
	}
	log.Printf("got events from %d to %d: %#v", fromMessage, messageIndex, events)

	chunk = types.NewEventStreamRange(events, start, end)

	return chunk, nil

	//	ch = eventMux.listen(user)
	//	to = eventBuffer.max // atomic

	//	resultChan = make(chan, limit)

	//	result = []
	//	found = 0
	//	event = indices[from]

	//	while found < limit && event.index < to:
	//		event = event.next
	//		if filter.test(user, event):
	//			resultChan <- event
	//			found += 1

	//	if found < limit {
	//		switch {
	//		new <- ch
	//			if new.index > to:

	//		default:
	//		}
	//		new = readAll(ch)

	//	}
}

func (s eventService) Messages(
	user types.UserId,
	room types.RoomId,
	from, to *types.StreamToken,
	limit uint,
) (eventRange *types.EventStreamRange, err types.Error) {
	maxMessage := s.messageSource.Max()

	var fromMessage uint64
	var presenceIndex uint64
	var typingIndex uint64

	if from != nil {
		fromMessage = from.MessageIndex
		presenceIndex = from.PresenceIndex
		typingIndex = from.TypingIndex
		if fromMessage > maxMessage {
			fromMessage = maxMessage
		}
	} else {
		fromMessage = maxMessage
		presenceIndex = s.presenceSource.Max()
		typingIndex = s.typingSource.Max()
	}

	var toMessage uint64

	if to != nil {
		toMessage = to.MessageIndex
	} else {
		toMessage = maxMessage
	}
	log.Println("to message", toMessage, to)

	roomSet := map[types.RoomId]struct{}{
		room: struct{}{},
	}

	messages, err := s.messageSource.Range(nil, nil, roomSet, fromMessage, toMessage, limit)
	if err != nil {
		return nil, err
	}

	log.Printf("getting messages from %d to %d, max %d", fromMessage, toMessage, maxMessage)

	messagesStart := fromMessage
	messagesEnd := fromMessage

	if len(messages) > 0 {
		messagesStart = messages[0].Index()
		messagesEnd = messages[len(messages)-1].Index() + 1
	}

	if to != nil && to.MessageIndex < fromMessage {
		messagesEnd, messagesStart = messagesStart, messagesEnd
	}

	start := types.NewStreamToken(messagesStart, presenceIndex, typingIndex)
	end := types.NewStreamToken(messagesEnd, presenceIndex, typingIndex)

	events := make([]types.Event, len(messages))

	for i, _ := range events {
		events[i] = messages[i].Event()
	}
	log.Printf("got messages from %d to %d: %#v", messagesStart, messagesEnd, events)

	eventRange = types.NewEventStreamRange(events, start, end)

	return eventRange, nil
}
