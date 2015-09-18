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

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/events"

	"github.com/Rugvip/bullettime/service"
	"github.com/Rugvip/bullettime/types"
	"github.com/julienschmidt/httprouter"

	"github.com/Rugvip/bullettime/api"
)

func setupApiEndpoint() http.Handler {
	roomStore, err := db.NewRoomDb()
	if err != nil {
		panic(err)
	}
	userStore, err := db.NewUserDb()
	if err != nil {
		panic(err)
	}
	aliasStore, err := db.NewAliasDb()
	if err != nil {
		panic(err)
	}
	memberStore, err := db.NewMembershipDb()
	if err != nil {
		panic(err)
	}
	streamMux, err := events.NewStreamMux()
	if err != nil {
		panic(err)
	}
	messageStream, err := events.NewMessageStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}
	presenceStream, err := events.NewPresenceStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}
	typingStream, err := events.NewTypingStream(memberStore, streamMux)
	if err != nil {
		panic(err)
	}

	roomService, err := service.CreateRoomService(
		roomStore,
		aliasStore,
		memberStore,
		messageStream,
		presenceStream,
		typingStream,
		typingStream,
	)
	if err != nil {
		panic(err)
	}
	userService, err := service.CreateUserService(userStore)
	if err != nil {
		panic(err)
	}
	profileService, err := service.NewProfileService(
		presenceStream,
		presenceStream,
		memberStore,
		roomStore,
		messageStream,
	)
	if err != nil {
		panic(err)
	}
	presenceService, err := service.NewPresenceService(presenceStream, presenceStream)
	if err != nil {
		panic(err)
	}
	tokenService, err := service.CreateTokenService()
	if err != nil {
		panic(err)
	}
	eventService, err := service.NewEventService(
		messageStream,
		presenceStream,
		typingStream,
		streamMux,
		messageStream,
		memberStore,
	)
	if err != nil {
		panic(err)
	}
	syncService, err := service.NewSyncService(
		messageStream,
		presenceStream,
		typingStream,
		roomStore,
		memberStore,
	)
	if err != nil {
		panic(err)
	}

	mux := httprouter.New()
	api.NewAuthEndpoint(userService, tokenService).Register(mux)
	api.NewProfileEndpoint(userService, tokenService, profileService).Register(mux)
	api.NewPresenceEndpoint(userService, tokenService, presenceService).Register(mux)
	api.NewRoomsEndpoint(userService, tokenService, roomService, syncService, eventService).Register(mux)
	api.NewEventsEndpoint(userService, tokenService, eventService, syncService).Register(mux)

	mux.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		api.WriteJsonResponseWithStatus(rw, types.DefaultUnrecognizedError)
	})

	mux.OPTIONS("/*path", func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	})

	corsHandler := http.NewServeMux()
	corsHandler.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		mux.ServeHTTP(rw, req)
	})

	return corsHandler
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/_matrix/client/api/v1/", http.StripPrefix("/_matrix/client/api/v1", setupApiEndpoint()))

	port := "4080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Listening on port " + port)
	log.Fatal(server.ListenAndServe())
}
