package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"github.com/julienschmidt/httprouter"
)

func (e eventsEndpoint) getEvents(req *http.Request) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}

	var from *types.StreamToken
	var to *types.StreamToken

	query := req.URL.Query()
	fromStr := query.Get("from")
	toStr := query.Get("to")
	dir := query.Get("dir")
	limitStr := query.Get("limit")
	timeoutStr := query.Get("timeout")

	if fromStr != "" {
		token, err := types.ParseStreamToken(fromStr)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		from = &token
	}

	if toStr != "" {
		token, err := types.ParseStreamToken(toStr)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		to = &token
	}

	if dir == "b" {
		token := types.NewStreamToken(0, 0, 0)
		to = &token
	}

	var limit uint64
	var timeout uint64

	var parseErr error
	if limitStr == "" {
		limit = 10 //TODO: make configurable
	} else {
		limit, parseErr = strconv.ParseUint(limitStr, 10, 32)
		if parseErr != nil {
			return types.BadQueryError(parseErr.Error())
		}
		if limit > 100 {
			limit = 100 //TODO: make configurable
		}
	}

	if timeoutStr == "" {
		timeout = 5000 //TODO: make configurable
	} else {
		timeout, parseErr = strconv.ParseUint(timeoutStr, 10, 32)
		if parseErr != nil {
			return types.BadQueryError(parseErr.Error())
		}
		if timeout > 60000 {
			timeout = 60000 //TODO: make configurable
		}
		if timeout < 100 {
			timeout = 100
		}
	}

	cancel := make(chan struct{})

	go func(timeout time.Duration) {
		time.Sleep(timeout)
		close(cancel)
	}(time.Millisecond * time.Duration(timeout))

	chunk, err := e.eventService.Range(authedUser, from, to, uint(limit), cancel)
	if err != nil {
		return err
	}

	return chunk
}

func (e eventsEndpoint) getSingleEvent(req *http.Request, params httprouter.Params) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	eventId, parseErr := types.ParseEventId(params[0].Value)
	if parseErr != nil {
		return types.BadJsonError(parseErr.Error())
	}
	event, err := e.eventService.Event(authedUser, eventId)
	if err != nil {
		return err
	}
	return event
}

func (e eventsEndpoint) getInitialSync(req *http.Request) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}

	query := req.URL.Query()
	limitStr := query.Get("limit")

	var limit uint64
	var parseErr error
	if limitStr == "" {
		limit = 10 //TODO: make configurable
	} else {
		limit, parseErr = strconv.ParseUint(limitStr, 10, 32)
		if parseErr != nil {
			return types.BadQueryError(parseErr.Error())
		}
		if limit > 100 {
			limit = 100 //TODO: make configurable
		}
	}

	initialSync, err := e.syncService.FullSync(authedUser, uint(limit))
	if err != nil {
		return err
	}
	return initialSync
}

func (e eventsEndpoint) getPublicRooms(req *http.Request) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	return authedUser
}

func (e eventsEndpoint) Register(mux *httprouter.Router) {
	mux.GET("/events", jsonHandler(e.getEvents))
	mux.PUT("/events/:eventId", jsonHandler(e.getSingleEvent))
	mux.GET("/initialSync", jsonHandler(e.getInitialSync))
	mux.PUT("/publicRooms", jsonHandler(e.getPublicRooms))
}

type eventsEndpoint struct {
	userService  interfaces.UserService
	tokenService interfaces.TokenService
	eventService interfaces.EventService
	syncService  interfaces.SyncService
}

func NewEventsEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	eventService interfaces.EventService,
	syncService interfaces.SyncService,
) Endpoint {
	return eventsEndpoint{
		userService,
		tokenService,
		eventService,
		syncService,
	}
}
