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

	mux := httprouter.New()
	api.NewAuthEndpoint(userService, tokenService).Register(mux)
	api.NewProfileEndpoint(userService, tokenService, profileService).Register(mux)
	api.NewRoomsEndpoint(userService, tokenService, roomService).Register(mux)
	api.NewEventsEndpoint(userService, tokenService, eventService).Register(mux)

	mux.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		api.WriteJsonResponseWithStatus(rw, types.DefaultUnrecognizedError)
	})

	return mux
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
