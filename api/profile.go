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
	profile, err := user.GetProfile()
	if err != nil {
		return err
	}
	return displayNameResponse{profile.DisplayName}
}

func setDisplayName(req *http.Request, params httprouter.Params, body *displayNameRequest) interface{} {
	authedUser, err := readAccessToken(req)
	if err != nil {
		return err
	}
	user, err := userFromParams(params)
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

func getAvatarUrl(params httprouter.Params) interface{} {
	user, err := userFromParams(params)
	if err != nil {
		return err
	}
	profile, err := user.GetProfile()
	if err != nil {
		return err
	}
	return avatarUrlResponse{profile.AvatarUrl}
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
	if body.AvatarUrl == nil {
		return types.BadJsonError("missing 'avatar_url'")
	}
	if err := user.SetAvatarUrl(*body.AvatarUrl, authedUser); err != nil {
		return err
	}
	return struct{}{}
}

func userFromParams(params httprouter.Params) (service.User, error) {
	userId, err := types.ParseUserId(params[0].Value)
	if err != nil {
		return service.User{}, types.BadJsonError(err.Error())
	}
	user, err := service.GetUser(userId)
	if err != nil {
		return service.User{}, err
	}
	return user, nil
}

func registerProfileResources(mux *httprouter.Router) {
	mux.GET("/profile/:userId/displayname", jsonHandler(getDisplayName))
	mux.PUT("/profile/:userId/displayname", jsonHandler(setDisplayName))
	mux.GET("/profile/:userId/avatar_url", jsonHandler(getAvatarUrl))
	mux.PUT("/profile/:userId/avatar_url", jsonHandler(setAvatarUrl))
	mux.GET("/user", jsonHandler(func() interface{} {
		user := new(types.User)
		user.UserId = types.NewUserId("test", "localhost")
		user.DisplayName = "Testan"
		user.AvatarUrl = "http://avatar.com"
		user.Presence = types.PresenceOnline
		user.LastActive = types.LastActive(time.Now())
		return user
	}))
	mux.POST("/user", jsonHandler(func(user *db.User) interface{} {
		log.Println("got user:", user)
		return user
	}))
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
