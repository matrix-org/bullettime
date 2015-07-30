package main

import (
	"log"
	"net/http"
	"os"
)

type testHandler struct{}

func (t testHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("Hello World"))
}

func main() {
	handler := testHandler{}
	mux := http.NewServeMux()
	mux.Handle("/", handler)

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
