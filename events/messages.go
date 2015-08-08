package events

import (
	"container/list"
	"sync"

	"github.com/Rugvip/bullettime/types"
)

type listItem struct {
	event types.Event
	index uint64
}

type messageSource struct {
	lock     sync.RWMutex
	list     *list.List
	elements map[types.Id]*list.Element
	indices  []*list.Element
	max      uint64
}

func NewMessageSource() (*messageSource, error) {
	return &messageSource{
		list:     list.New(),
		elements: map[types.Id]*list.Element{},
		indices:  []*list.Element{},
		max:      0,
	}, nil
}

func (h *messageSource) Send(event *types.Message) (uint64, types.Error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	index := h.max
	h.max += 1
	item := listItem{event, index}

	element := h.elements[event.Id()]
	if element != nil {
		h.indices[element.Value.(listItem).index] = element.Next()
		h.list.MoveToFront(element)
		element.Value = item
	} else {
		element = h.list.PushFront(listItem{event, index})
		h.elements[event.Id()] = element
	}
	h.indices = append(h.indices, element)
	return index, nil
}

func (h *messageSource) Iterate(index uint64) ([]types.Event, types.Error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	result := []types.Event{}
	e := h.list.Front()
	for e != nil && e.Value.(listItem).index > index {
		result = append(result, e.Value.(listItem).event)
		e = e.Next()
	}
	return result, nil
}

func (h *messageSource) Max() (index uint64, err types.Error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.max, nil
}
