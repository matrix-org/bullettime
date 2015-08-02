package api

import (
	"log"
	"net/http"
	"time"

	"github.com/Rugvip/bullettime/db"
	"github.com/Rugvip/bullettime/service"
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

func getDisplayName(params httprouter.Params) interface{} {
	user, err := userFromParams(params)
	if err != nil {
		return err
	}
	name, err := user.GetDisplayName()
	if err != nil {
		return ServerError(err.Error())
	}
	return displayNameResponse{name}
}

func setDisplayName(req *http.Request, params httprouter.Params, body *displayNameRequest) interface{} {
	authedUser, apiErr := readAccessToken(req)
	if apiErr != nil {
		return apiErr
	}
	user, err := userFromParams(params)
	if err != nil {
		return err
	}
	if authedUser.Id() != user.Id() {
		return ForbiddenError("can't change the display name of other users")
	}
	if body.DisplayName == nil {
		return BadJsonError("missing 'displayname'")
	}
	if err := user.SetDisplayName(*body.DisplayName); err != nil {
		return ServerError(err.Error())
	}
	return struct{}{}
}

func getAvatarUrl(params httprouter.Params) interface{} {
	user, err := userFromParams(params)
	if err != nil {
		return err
	}
	url, err := user.GetAvatarUrl()
	if err != nil {
		return ServerError(err.Error())
	}
	return avatarUrlResponse{url}
}

func setAvatarUrl(req *http.Request, params httprouter.Params, body *avatarUrlRequest) interface{} {
	authedUser, err := readAccessToken(req)
	if err != nil {
		return err
	}
	user, err := userFromParams(params)
	if err != nil {
		return err
	}
	if authedUser.Id() != user.Id() {
		return ForbiddenError("can't change the avatar url of other users")
	}
	if body.AvatarUrl == nil {
		return BadJsonError("missing 'avatar_url'")
	}
	if err := user.SetAvatarUrl(*body.AvatarUrl); err != nil {
		return ServerError(err.Error())
	}
	return struct{}{}
}

func userFromParams(params httprouter.Params) (service.User, error) {
	userId, err := types.ParseUserId(params[0].Value)
	if err != nil {
		return service.User{}, BadJsonError(err.Error())
	}
	user, err := service.GetUser(userId)
	if err != nil {
		return service.User{}, NotFoundError(err.Error())
	}
	return user, nil
}

func registerProfileResources(mux *httprouter.Router) {
	mux.GET("/profile/:userId/displayname", jsonHandler(getDisplayName))
	mux.PUT("/profile/:userId/displayname", jsonHandler(setDisplayName))
	mux.GET("/profile/:userId/avatar_url", jsonHandler(getAvatarUrl))
	mux.PUT("/profile/:userId/avatar_url", jsonHandler(setAvatarUrl))
	mux.GET("/user/:userId", jsonHandler(func(params httprouter.Params) interface{} {
		userId, err := types.ParseUserId(params[0].Value)
		if err != nil {
			return BadJsonError(err.Error())
		}
		log.Printf("get user %s", userId)
		return db.GetUser(userId)
	}))
	mux.GET("/user", jsonHandler(func() interface{} {
		return &db.User{
			Id: types.NewUserId("test", "localhost"),
			UserProfile: types.UserProfile{
				DisplayName: "Testan",
				AvatarUrl:   "http://avatar.com",
			},
			UserPresence: types.UserPresence{
				Presence:   types.PresenceOnline,
				LastActive: types.LastActive(time.Now()),
			},
		}
	}))
	mux.POST("/user", jsonHandler(func(user *db.User) interface{} {
		log.Println("got user:", user)
		return user
	}))
	mux.GET("/event", jsonHandler(func() interface{} {
		return &types.Event{
			Id:        types.NewEventId("dkjfhg", "localhost"),
			RoomId:    types.NewRoomId("dkfghu", "localhost"),
			EventType: "m.test",
			Timestamp: types.Timestamp{time.Now()},
			Content: types.TestContent{
				Name: "test",
			},
		}
	}))
	mux.POST("/event/:eventType/:stateKey", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := types.State{
			Event: types.Event{
				Id:        types.NewEventId("123", "localhost"),
				RoomId:    types.NewRoomId("abc", "localhost"),
				EventType: types.EventType(params[0].Value),
				Timestamp: types.Timestamp{time.Now()},
				Content:   content,
			},
			StateKey: params[1].Value,
		}
		log.Println("got state: ", event)
		return &event
	}))
	mux.POST("/event/:eventType/", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := types.State{
			Event: types.Event{
				Id:        types.NewEventId("123", "localhost"),
				RoomId:    types.NewRoomId("abc", "localhost"),
				EventType: types.EventType(params[0].Value),
				Timestamp: types.Timestamp{time.Now()},
				Content:   content,
			},
			StateKey: "",
		}
		log.Println("got state: ", event)
		return &event
	}))
	mux.POST("/event/:eventType", jsonHandler(func(params httprouter.Params, content *types.TestContent) interface{} {
		event := types.Event{
			Id:        types.NewEventId("123", "localhost"),
			RoomId:    types.NewRoomId("abc", "localhost"),
			Content:   content,
			EventType: types.EventType(params[0].Value),
			Timestamp: types.Timestamp{time.Now()},
		}
		log.Println("got event: ", event)
		return &event
	}))
}
