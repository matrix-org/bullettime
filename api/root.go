package api

import (
	"log"

	"github.com/julienschmidt/httprouter"

	"net/http"
)

func NewRootMux() http.Handler {
	mux := httprouter.New()
	mux.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		writeJsonResponseWithStatus(rw, defaultUnrecognizedError)
	})
	mux.PanicHandler = func(rw http.ResponseWriter, req *http.Request, object interface{}) {
		log.Println("Request to "+req.URL.Path+" ended in panic:", object)
		writeJsonResponseWithStatus(rw, ServerError("internal server error"))
	}
	registerAuthResources(mux)
	registerProfileResources(mux)
	return mux
}
