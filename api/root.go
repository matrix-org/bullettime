package api

import (
	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"github.com/julienschmidt/httprouter"

	"net/http"
)

func NewRootMux(roomService interfaces.RoomService) http.Handler {
	mux := httprouter.New()
	mux.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		writeJsonResponseWithStatus(rw, types.DefaultUnrecognizedError)
	})
	// mux.PanicHandler = func(rw http.ResponseWriter, req *http.Request, object interface{}) {
	// 	log.Println("Request to "+req.URL.Path+" ended in panic:", object, reflect.TypeOf(object))
	// 	writeJsonResponseWithStatus(rw, types.ServerError("internal server error"))
	// }
	registerAuthResources(mux)
	registerProfileResources(mux)
	NewRoomsEndpoint(roomService).register(mux)
	return mux
}
