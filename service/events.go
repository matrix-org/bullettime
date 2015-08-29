package service

import (
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
		return nil, types.NotFoundError("event not found: " + event.GetEventId().String())
	}
	return event, nil
}

func (s eventService) Range(
	user types.UserId,
	from, to *types.StreamToken,
	limit uint,
	cancel chan struct{},
) (chunk *types.EventStreamChunk, err types.Error) {
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
		fromPresence = from.PresenceIndex
		fromTyping = from.TypingIndex
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

	messages, err := s.messageSource.Range(user, userSet, roomSet, fromMessage, toMessage, limit)
	if err != nil {
		return nil, err
	}
	presences, err := s.presenceSource.Range(user, userSet, roomSet, fromPresence, toPresence, limit)
	if err != nil {
		return nil, err
	}
	typings, err := s.typingSource.Range(user, userSet, roomSet, fromTyping, toTyping, limit)
	if err != nil {
		return nil, err
	}

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

	chunk = types.NewEventStreamChunk(events, start, end)

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
