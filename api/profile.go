package api

import (
	"log"
	"net/http"

	"github.com/Rugvip/bullettime/service"

	"github.com/julienschmidt/httprouter"
)

type avatarUrlBody struct {
	AvatarUrl string `json:"avatar_url"`
}
type displayNameBody struct {
	DisplayName string `json:"displayname"`
}

func getDisplayName(params httprouter.Params) interface{} {
	user, err := service.GetUser(params[0].Value)
	if err != nil {
		return NotFoundError(err.Error())
	}
	name, err := user.GetDisplayName()
	if err != nil {
		return ServerError(err.Error())
	}
	return displayNameBody{name}
}

func setDisplayName(req *http.Request, params httprouter.Params, body *displayNameBody) interface{} {
	log.Println("set display name", body)
	authedUser, apiErr := readAccessToken(req)
	if apiErr != nil {
		return apiErr
	}
	user, err := service.GetUser(params[0].Value)
	if err != nil {
		return NotFoundError(err.Error())
	}
	if authedUser.Id() != user.Id() {
		return ForbiddenError("can't change the display name of other users")
	}
	if err := user.SetDisplayName(body.DisplayName); err != nil {
		return ServerError(err.Error())
	}
	return struct{}{}
}

func getAvatarUrl(params httprouter.Params) interface{} {
	user, err := service.GetUser(params[0].Value)
	if err != nil {
		return NotFoundError(err.Error())
	}
	url, err := user.GetAvatarUrl()
	if err != nil {
		return ServerError(err.Error())
	}
	return avatarUrlBody{url}
}

func setAvatarUrl(req *http.Request, params httprouter.Params, body *avatarUrlBody) interface{} {
	authedUser, apiErr := readAccessToken(req)
	if apiErr != nil {
		return apiErr
	}
	user, err := service.GetUser(params[0].Value)
	if err != nil {
		return NotFoundError(err.Error())
	}
	if authedUser.Id() != user.Id() {
		return ForbiddenError("can't change the avatar url of other users")
	}
	if err := user.SetAvatarUrl(body.AvatarUrl); err != nil {
		return ServerError(err.Error())
	}
	return struct{}{}
}

func registerProfileResources(mux *httprouter.Router) {
	mux.GET("/profile/:userId/displayname", jsonHandler(getDisplayName))
	mux.PUT("/profile/:userId/displayname", jsonHandler(setDisplayName))
	mux.GET("/profile/:userId/avatar_url", jsonHandler(getAvatarUrl))
	mux.PUT("/profile/:userId/avatar_url", jsonHandler(setAvatarUrl))
}
