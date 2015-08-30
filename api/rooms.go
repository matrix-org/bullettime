package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

func (e roomsEndpoint) doWildcardJoin(req *http.Request, params httprouter.Params) interface{} {
	user, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	roomIdOrAlias := params[0].Value
	room, parseErr := types.ParseRoomId(roomIdOrAlias)
	if parseErr != nil {
		alias, parseErr := types.ParseAlias(roomIdOrAlias)
		if parseErr != nil {
			return types.BadParamError("invalid room id or alias: " + roomIdOrAlias)
		}
		room, err = e.roomService.LookupAlias(alias)
		if err != nil {
			return err
		}
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipMember
	_, err = e.roomService.SetState(room, user, &content, user.String())
	if err != nil {
		return err
	}
	return struct{}{}
}

func (e roomsEndpoint) sendMessage(req *http.Request, params httprouter.Params, content *map[string]interface{}) interface{} {
	user, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	room, parseErr := types.ParseRoomId(params[0].Value)
	if parseErr != nil {
		return types.BadParamError(parseErr.Error())
	}
	eventType := params[1].Value
	typedContent := types.NewGenericContent(*content, eventType)
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

func (e roomsEndpoint) doKick(req *http.Request, params httprouter.Params, body *userRequest) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipLeaving
	state, err := e.roomService.SetState(room, user, &content, body.UserId.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) doBan(req *http.Request, params httprouter.Params, body *userRequest) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipBanned
	state, err := e.roomService.SetState(room, user, &content, body.UserId.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) doJoin(req *http.Request, params httprouter.Params) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipMember
	state, err := e.roomService.SetState(room, user, &content, user.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) doKnock(req *http.Request, params httprouter.Params) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipKnocking
	state, err := e.roomService.SetState(room, user, &content, user.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) doLeave(req *http.Request, params httprouter.Params) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}
	content := types.MembershipEventContent{}
	content.Membership = types.MembershipLeaving
	state, err := e.roomService.SetState(room, user, &content, user.String())
	if err != nil {
		return err
	}
	return eventIdResponse{state.EventId}
}

func (e roomsEndpoint) doInitialSync(req *http.Request, params httprouter.Params) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}

	query := req.URL.Query()
	limitStr := query.Get("limit")
	var limit uint64

	var parseErr error
	if limitStr == "" {
		limit = 10 //TODO: make configurable
	} else {
		limit, parseErr = strconv.ParseUint(limitStr, 10, 32)
		if parseErr != nil {
			return types.BadQueryError(parseErr.Error())
		}
		if limit > 100 {
			limit = 100 //TODO: make configurable
		}
	}

	roomSync, err := e.syncService.RoomSync(user, room, uint(limit))
	if err != nil {
		return err
	}
	return roomSync
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
		jsonErr = json.NewDecoder(req.Body).Decode(&genericContent.Content)
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

func (e roomsEndpoint) getMessages(req *http.Request, params httprouter.Params) interface{} {
	room, user, err := e.getRoomAndUser(req, params)
	if err != nil {
		return err
	}

	var from *types.StreamToken
	var to *types.StreamToken

	query := req.URL.Query()
	fromStr := query.Get("from")
	toStr := query.Get("to")
	dir := query.Get("dir")
	limitStr := query.Get("limit")

	if fromStr != "" {
		token, err := types.ParseStreamToken(fromStr)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		from = &token
	}

	if toStr != "" {
		token, err := types.ParseStreamToken(toStr)
		if err != nil {
			return types.BadQueryError(err.Error())
		}
		to = &token
	}

	if dir == "b" {
		token := types.NewStreamToken(0, 0, 0)
		to = &token
	}

	var limit uint64

	var parseErr error
	if limitStr == "" {
		limit = 10 //TODO: make configurable
	} else {
		limit, parseErr = strconv.ParseUint(limitStr, 10, 32)
		if parseErr != nil {
			return types.BadQueryError(parseErr.Error())
		}
		if limit > 100 {
			limit = 100 //TODO: make configurable
		}
	}
	eventRange, err := e.eventService.Messages(user, room, from, to, uint(limit))
	log.Println("TO", to, eventRange)
	if err != nil {
		return err
	}

	return eventRange
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
	mux.PUT("/rooms/:roomId/send/:eventType/:txn", jsonHandler(e.sendMessage))
	// mux.GET("/rooms/:roomId/state/:eventType", jsonHandler(dummy))
	mux.PUT("/rooms/:roomId/state/:eventType", e.handlePutState)
	mux.PUT("/rooms/:roomId/state/:eventType/:stateKey", e.handlePutState)
	// mux.GET("/rooms/:roomId/state/:eventType/:stateKey", jsonHandler(dummy))
	mux.POST("/rooms/:roomId/invite", jsonHandler(e.doInvite))
	mux.POST("/rooms/:roomId/kick", jsonHandler(e.doKick))
	mux.POST("/rooms/:roomId/ban", jsonHandler(e.doBan))
	mux.POST("/rooms/:roomId/join", jsonHandler(e.doJoin))
	mux.POST("/rooms/:roomId/knock", jsonHandler(e.doKnock))
	mux.POST("/rooms/:roomId/leave", jsonHandler(e.doLeave))
	mux.GET("/rooms/:roomId/messages", jsonHandler(e.getMessages))
	// mux.GET("/rooms/:roomId/members", jsonHandler(dummy))
	// mux.GET("/rooms/:roomId/state", jsonHandler(dummy))
	// mux.PUT("/rooms/:roomId/typing/:userId", jsonHandler(dummy))
	mux.GET("/rooms/:roomId/initialSync", jsonHandler(e.doInitialSync))
	mux.POST("/join/:roomAliasOrId", jsonHandler(e.doWildcardJoin))
	mux.POST("/createRoom", jsonHandler(e.createRoom))
}

type roomsEndpoint struct {
	userService  interfaces.UserService
	tokenService interfaces.TokenService
	roomService  interfaces.RoomService
	syncService  interfaces.SyncService
	eventService interfaces.EventService
}

func NewRoomsEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	roomService interfaces.RoomService,
	syncService interfaces.SyncService,
	eventService interfaces.EventService,
) Endpoint {
	return roomsEndpoint{
		userService,
		tokenService,
		roomService,
		syncService,
		eventService,
	}
}
