// Copyright 2015  Ericsson AB
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

package api

import (
	"net/http"
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

	query := urlQuery{req.URL.Query()}

	from, err := query.parseStreamToken("from")
	if err != nil {
		return err
	}

	to, err := query.parseStreamToken("to")
	if err != nil {
		return err
	}

	limit, err := query.parseUint("limit", 10)
	if err != nil {
		return err
	}
	if limit > 100 {
		limit = 100 //TODO: make configurable
	}

	timeout, err := query.parseUint("timeout", 5000)
	if err != nil {
		return err
	}
	if timeout > 60000 {
		timeout = 60000 //TODO: make configurable
	}
	if timeout < 100 {
		timeout = 100
	}

	dir := query.Get("dir")
	if dir == "b" {
		token := types.NewStreamToken(0, 0, 0)
		to = &token
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

	query := urlQuery{req.URL.Query()}

	limit, err := query.parseUint("limit", 10)
	if err != nil {
		return err
	}
	if limit > 100 {
		limit = 100 //TODO: make configurable
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
