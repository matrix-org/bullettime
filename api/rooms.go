package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Rugvip/bullettime/interfaces"

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

type eventIdResponse struct {
	EventId types.EventId `json:"event_id"`
}

type userRequest struct {
	UserId types.UserId `json:"user_id"`
}

func (e roomsEndpoint) createRoom(req *http.Request, body *types.RoomDescription) interface{} {
	creator, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	hostname := strings.Split(req.Host, ":")[0]
	room, alias, err := e.roomService.CreateRoom(hostname, creator, body)
	if err != nil {
		return err
	}
	return CreateRoomResponse{room, alias}
}

func (e roomsEndpoint) sendMessage(req *http.Request, params httprouter.Params, content map[string]interface{}) interface{} {
	user, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	room, parseErr := types.ParseRoomId(params[0].Value)
	if parseErr != nil {
		return types.BadParamError(parseErr.Error())
	}
	eventType := params[1].Value
	typedContent := types.NewGenericContent(content, eventType)
	message, err := e.roomService.AddMessage(room, user, typedContent)
	if err != nil {
		return err
	}
	return eventIdResponse{message.EventId}
}

func (e roomsEndpoint) doInvite(req *http.Request, params httprouter.Params, body *userRequest) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipInvited
	state, err := e.roomService.SetState(room, user, &content, body.UserId.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) handlePutState(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		WriteJsonResponseWithStatus(rw, err)
		return
	}
	eventType := params[1].Value
	stateKey := ""
	if len(params) > 2 {
		stateKey = params[2].Value
	}

	var content types.TypedContent
	switch eventType {
	case types.EventTypeMembership:
		content = &types.MembershipEventContent{}
	case types.EventTypeName:
		content = &types.NameEventContent{}
	case types.EventTypeTopic:
		content = &types.TopicEventContent{}
	case types.EventTypePowerLevels:
		content = &types.PowerLevelsEventContent{}
	case types.EventTypeJoinRules:
		content = &types.JoinRulesEventContent{}
	}
	var jsonErr error
	if content != nil {
		jsonErr = json.NewDecoder(req.Body).Decode(content)
	} else {
		genericContent := types.NewGenericContent(map[string]interface{}{}, eventType)
		content = genericContent
		jsonErr = json.NewDecoder(req.Body).Decode(genericContent.Content)
	}
	if jsonErr != nil {
		switch err := jsonErr.(type) {
		case *json.SyntaxError:
			msg := fmt.Sprintf("error at [%d]: %s", err.Offset, err.Error())
			WriteJsonResponseWithStatus(rw, types.NotJsonError(msg))
		case *json.UnmarshalTypeError:
			msg := fmt.Sprintf("error at [%d]: expected type %s but got %s", err.Offset, err.Type, err.Value)
			WriteJsonResponseWithStatus(rw, types.BadJsonError(msg))
		default:
			WriteJsonResponseWithStatus(rw, types.BadJsonError(err.Error()))
		}
		return
	}
	state, err := e.roomService.SetState(room, user, content, stateKey)
	if err != nil {
		WriteJsonResponseWithStatus(rw, err)
		return
	}
	res := eventIdResponse{state.EventId}
	WriteJsonResponse(rw, 200, res)
}

func (e roomsEndpoint) getRoomAndUser(req *http.Request, params httprouter.Params) (types.RoomId, types.UserId, types.Error) {
	user, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return types.RoomId{}, types.UserId{}, err
	}
	room, parseErr := types.ParseRoomId(params[0].Value)
	if parseErr != nil {
		return types.RoomId{}, types.UserId{}, types.BadParamError(parseErr.Error())
	}
	return room, user, nil
}

func (e roomsEndpoint) Register(mux *httprouter.Router) {
	mux.POST("/rooms/:roomId/send/:eventType", jsonHandler(e.sendMessage))
	// mux.GET("/rooms/:roomId/state/:eventType", jsonHandler(dummy))
	mux.PUT("/rooms/:roomId/state/:eventType", e.handlePutState)
	mux.PUT("/rooms/:roomId/state/:eventType/:stateKey", e.handlePutState)
	// mux.GET("/rooms/:roomId/state/:eventType/:stateKey", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/invite", jsonHandler(e.doInvite))
	// mux.POST("/rooms/:roomId/join", jsonHandler(dummy))
	// mux.POST("/rooms/:roomId/leave", jsonHandler(dummy))
	// mux.POST("/rooms/:roomId/ban", jsonHandler(dummy))
	// mux.GET("/rooms/:roomId/messages", jsonHandler(dummy))
	// mux.GET("/rooms/:roomId/members", jsonHandler(dummy))
	// mux.GET("/rooms/:roomId/state", jsonHandler(dummy))
	// mux.PUT("/rooms/:roomId/typing/:userId", jsonHandler(dummy))
	// mux.GET("/rooms/:roomId/initialSync", jsonHandler(dummy))
	// mux.POST("/join/:roomAliasOrId", jsonHandler(dummy))
	mux.POST("/createRoom", jsonHandler(e.createRoom))
}

type roomsEndpoint struct {
	userService  interfaces.UserService
	tokenService interfaces.TokenService
	roomService  interfaces.RoomService
}

func NewRoomsEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	roomService interfaces.RoomService,
) Endpoint {
	return roomsEndpoint{
		userService:  userService,
		tokenService: tokenService,
		roomService:  roomService,
	}
}
