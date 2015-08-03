package api

import (
	"net/http"
	"strings"

	"github.com/Rugvip/bullettime/service"

	"github.com/Rugvip/bullettime/types"

	"github.com/julienschmidt/httprouter"
)

func dummy() interface{} {
	return struct{}{}
}

type CreateRoomResponse struct {
	RoomId    types.RoomId `json:"room_id"`
	RoomAlias *types.Alias `json:"room_alias,omitempty"`
}

func createRoom(req *http.Request, body *types.RoomDescription) interface{} {
	creator, err := readAccessToken(req)
	if err != nil {
		return err
	}
	hostname := strings.Split(req.Host, ":")[0]
	roomId, alias, err := service.CreateRoom(hostname, creator, body)
	if err != nil {
		return err
	}
	return CreateRoomResponse{roomId, alias}
}

func registerRoomResources(mux *httprouter.Router) {
	mux.POST("/rooms/:roomId/send/:eventType", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/state/:eventType", jsonHandler(dummy))
	mux.PUT("/rooms/:roomId/state/:eventType", jsonHandler(dummy))
	mux.PUT("/rooms/:roomId/state/:eventType/:stateKey", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/state/:eventType/:stateKey", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/invite", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/join", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/leave", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/ban", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/messages", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/members", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/state", jsonHandler(dummy))
	mux.PUT("/rooms/:roomId/typing/:userId", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/initialSync", jsonHandler(dummy))
	mux.POST("/join/:roomAliasOrId", jsonHandler(dummy))
	mux.POST("/createRoom", jsonHandler(createRoom))
}
