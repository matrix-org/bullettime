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

	if limitStr == "" {
		limit = 10 //TODO: make configurable
	} else {
		limit, err = strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		if limit > 100 {
			limit = 100 //TODO: make configurable
		}
	}

	if timeoutStr == "" {
		timeout = 5000 //TODO: make configurable
	} else {
		timeout, err = strconv.ParseUint(timeoutStr, 10, 32)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		if timeout > 60000 {
			timeout = 60000 //TODO: make configurable
		}
	}

	cancel := make(chan struct{})

	go func() {
		time.Sleep(time.Millisecond * time.Duration(timeout))
		close(cancel)
	}()

	chunk, err := e.eventService.Range(authedUser.Id(), from, to, uint(limit), cancel)
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
	eventId, err := types.ParseEventId(params[0].Value)
	if err != nil {
		return types.BadJsonError(err.Error())
	}
	event, err := e.eventService.Event(authedUser.Id(), eventId)
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
	return authedUser
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
}

func NewEventsEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	eventService interfaces.EventService,
) Endpoint {
	return eventsEndpoint{
		userService:  userService,
		tokenService: tokenService,
		eventService: eventService,
	}
}
