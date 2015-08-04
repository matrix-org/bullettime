package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Rugvip/bullettime/service"

	"github.com/Rugvip/bullettime/api"
)

func main() {
	roomService, err := service.CreateRoomService()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api.NewRootMux(roomService)))

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
