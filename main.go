package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Rugvip/bullettime/service"
	"github.com/Rugvip/bullettime/types"
	"github.com/julienschmidt/httprouter"

	"github.com/Rugvip/bullettime/api"
)

func setupApiEndpoint() http.Handler {
	roomService, err := service.CreateRoomService()
	if err != nil {
		panic(err)
	}
	userService, err := service.CreateUserService()
	if err != nil {
		panic(err)
	}
	tokenService, err := service.CreateTokenService()
	if err != nil {
		panic(err)
	}

	mux := httprouter.New()
	api.NewAuthEndpoint(userService, tokenService).Register(mux)
	api.NewProfileEndpoint(userService, tokenService, roomService).Register(mux)
	api.NewRoomsEndpoint(userService, tokenService, roomService).Register(mux)

	mux.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		res, err := json.Marshal(types.DefaultUnrecognizedError)
		if err != nil {
			rw.WriteHeader(500)
			log.Println("marshaling error: ", err)
			fmt.Fprintf(rw, "{\"errcode\":\"M_SERVER_ERROR\",\"error\":\"failed to marshal response\"}")
		} else {
			rw.WriteHeader(types.DefaultUnrecognizedError.Status())
			rw.Write(res)
		}
	})

	return mux
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", setupApiEndpoint()))

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
