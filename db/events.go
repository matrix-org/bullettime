package db

import (
	"container/list"
	"sync"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
)

type listItem struct {
	event types.Event
	index uint64
}

type eventQueue struct {
	sync.RWMutex
	list     *list.List
	elements map[types.Id]*list.Element
	max      uint64
}

func NewEventStream() (interfaces.EventStream, error) {
	return &eventQueue{
		list:     list.New(),
		elements: map[types.Id]*list.Element{},
		max:      1,
	}, nil
}

func (h *eventQueue) Push(event types.Event) (uint64, types.Error) {
	h.Lock()
	defer h.Unlock()

	index := h.max
	h.max += 1
	item := listItem{event, index}

	if element := h.elements[event.Id()]; element != nil {
		h.list.MoveToFront(element)
		element.Value = item
	} else {
		element = h.list.PushFront(listItem{event, h.max})
		h.elements[event.Id()] = element
	}
	return index, nil
}

func (h *eventQueue) Iterate(index uint64) ([]types.Event, types.Error) {
	h.RLock()
	defer h.RUnlock()
	result := []types.Event{}
	e := h.list.Front()
	for e != nil && e.Value.(listItem).index > index {
		result = append(result, e.Value.(listItem).event)
		e = e.Next()
	}
	return result, nil
}

func (h *eventQueue) Max() (index uint64, err types.Error) {
	h.RLock()
	defer h.RUnlock()
	return h.max, nil
}
