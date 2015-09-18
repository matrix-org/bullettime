// Copyright 2015  Ericsson AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"net/http"

	"github.com/matrix-org/bullettime/interfaces"
	"github.com/matrix-org/bullettime/types"

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
	user, err := urlParams{params}.user(0, e.users)
	if err != nil {
		return err
	}
	profile, err := e.profiles.Profile(user, user)
	if err != nil {
		return err
	}
	return displayNameResponse{profile.DisplayName}
}

func (e profileEndpoint) setDisplayName(req *http.Request, params httprouter.Params, body *displayNameRequest) interface{} {
	authedUser, err := readAccessToken(e.users, e.tokens, req)
	if err != nil {
		return err
	}
	user, err := urlParams{params}.user(0, e.users)
	if err != nil {
		return err
	}
	if body.DisplayName == nil {
		return types.BadJsonError("missing 'displayname'")
	}
	if _, err := e.profiles.UpdateProfile(user, authedUser, body.DisplayName, nil); err != nil {
		return err
	}
	return struct{}{}
}

func (e profileEndpoint) getAvatarUrl(params httprouter.Params) interface{} {
	user, err := urlParams{params}.user(0, e.users)
	if err != nil {
		return err
	}
	profile, err := e.profiles.Profile(user, user)
	if err != nil {
		return err
	}
	return avatarUrlResponse{profile.AvatarUrl}
}

func (e profileEndpoint) setAvatarUrl(req *http.Request, params httprouter.Params, body *avatarUrlRequest) interface{} {
	authedUser, err := readAccessToken(e.users, e.tokens, req)
	if err != nil {
		return err
	}
	user, err := urlParams{params}.user(0, e.users)
	if err != nil {
		return err
	}
	if body.AvatarUrl == nil {
		return types.BadJsonError("missing 'avatar_url'")
	}
	if _, err := e.profiles.UpdateProfile(user, authedUser, nil, body.AvatarUrl); err != nil {
		return err
	}
	return struct{}{}
}

func (e profileEndpoint) getProfile(params httprouter.Params) interface{} {
	user, err := urlParams{params}.user(0, e.users)
	if err != nil {
		return err
	}
	profile, err := e.profiles.Profile(user, user)
	if err != nil {
		return err
	}
	return profile
}

func (e profileEndpoint) Register(mux *httprouter.Router) {
	mux.GET("/profile/:userId/displayname", jsonHandler(e.getDisplayName))
	mux.PUT("/profile/:userId/displayname", jsonHandler(e.setDisplayName))
	mux.GET("/profile/:userId/avatar_url", jsonHandler(e.getAvatarUrl))
	mux.PUT("/profile/:userId/avatar_url", jsonHandler(e.setAvatarUrl))
	mux.GET("/profile/:userId", jsonHandler(e.getProfile))
}

type profileEndpoint struct {
	users    interfaces.UserService
	tokens   interfaces.TokenService
	profiles interfaces.ProfileService
}

func NewProfileEndpoint(
	users interfaces.UserService,
	tokens interfaces.TokenService,
	profiles interfaces.ProfileService,
) Endpoint {
	return profileEndpoint{
		users,
		tokens,
		profiles,
	}
}
