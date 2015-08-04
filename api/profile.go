package api

import (
	"log"
	"net/http"
	"time"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"

	"github.com/julienschmidt/httprouter"
)

type avatarUrlRequest struct {
	AvatarUrl *string `json:"avatar_url"`
}
type avatarUrlResponse struct {
	AvatarUrl string `json:"avatar_url"`
}
type displayNameRequest struct {
	DisplayName *string `json:"displayname"`
}
type displayNameResponse struct {
	DisplayName string `json:"displayname"`
}

func (e profileEndpoint) getDisplayName(params httprouter.Params) interface{} {
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	profile, err := user.Profile()
	if err != nil {
		return err
	}
	return displayNameResponse{profile.DisplayName}
}

func (e profileEndpoint) setDisplayName(req *http.Request, params httprouter.Params, body *displayNameRequest) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	if body.DisplayName == nil {
		return types.BadJsonError("missing 'displayname'")
	}
	if err := user.SetDisplayName(*body.DisplayName, authedUser); err != nil {
		return err
	}
	return struct{}{}
}

func (e profileEndpoint) getAvatarUrl(params httprouter.Params) interface{} {
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	profile, err := user.Profile()
	if err != nil {
		return err
	}
	return avatarUrlResponse{profile.AvatarUrl}
}

func (e profileEndpoint) setAvatarUrl(req *http.Request, params httprouter.Params, body *avatarUrlRequest) interface{} {
	authedUser, err := readAccessToken(e.userService, e.tokenService, req)
	if err != nil {
		return err
	}
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	if body.AvatarUrl == nil {
		return types.BadJsonError("missing 'avatar_url'")
	}
	if err := user.SetAvatarUrl(*body.AvatarUrl, authedUser); err != nil {
		return err
	}
	return struct{}{}
}

func (e profileEndpoint) userFromParams(params httprouter.Params) (interfaces.User, error) {
	userId, err := types.ParseUserId(params[0].Value)
	if err != nil {
		return nil, types.BadJsonError(err.Error())
	}
	user, err := e.userService.User(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (e profileEndpoint) Register(mux *httprouter.Router) {
	mux.GET("/profile/:userId/displayname", jsonHandler(e.getDisplayName))
	mux.PUT("/profile/:userId/displayname", jsonHandler(e.setDisplayName))
	mux.GET("/profile/:userId/avatar_url", jsonHandler(e.getAvatarUrl))
	mux.PUT("/profile/:userId/avatar_url", jsonHandler(e.setAvatarUrl))
	mux.GET("/event", jsonHandler(func() interface{} {
		event := new(types.Message)
		event.EventId = types.NewEventId("dkjfhg", "localhost")
		event.RoomId = types.NewRoomId("dkfghu", "localhost")
		event.UserId = types.NewUserId("test", "localhost")
		event.EventType = "m.test"
		event.Timestamp = types.Timestamp{time.Now()}
		event.Content = types.TestContent{
			Name: "test",
		}
		return event
	}))
	mux.POST("/event/:eventType/:stateKey", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := new(types.State)
		event.EventId = types.NewEventId("123", "localhost")
		event.RoomId = types.NewRoomId("abc", "localhost")
		event.UserId = types.NewUserId("test", "localhost")
		event.EventType = params[0].Value
		event.Timestamp = types.Timestamp{time.Now()}
		event.Content = content
		event.StateKey = params[1].Value
		log.Println("got state: ", event)
		return event
	}))
	mux.POST("/event/:eventType/", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := new(types.State)
		event.EventId = types.NewEventId("123", "localhost")
		event.RoomId = types.NewRoomId("abc", "localhost")
		event.UserId = types.NewUserId("test", "localhost")
		event.EventType = params[0].Value
		event.Timestamp = types.Timestamp{time.Now()}
		event.Content = content
		event.StateKey = ""
		log.Println("got state: ", event)
		return event
	}))
	mux.POST("/event/:eventType", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := new(types.Message)
		event.EventId = types.NewEventId("123", "localhost")
		event.RoomId = types.NewRoomId("abc", "localhost")
		event.UserId = types.NewUserId("test", "localhost")
		event.Content = content
		event.EventType = params[0].Value
		event.Timestamp = types.Timestamp{time.Now()}
		log.Println("got event: ", event)
		return event
	}))
}

type profileEndpoint struct {
	userService  interfaces.UserService
	tokenService interfaces.TokenService
	roomService  interfaces.RoomService
}

func NewProfileEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	roomService interfaces.RoomService,
) Endpoint {
	return profileEndpoint{
		userService:  userService,
		tokenService: tokenService,
		roomService:  roomService,
	}
}
