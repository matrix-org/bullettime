// Copyright 2015 OpenMarket Ltd
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
	"fmt"
	"strings"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"github.com/Rugvip/bullettime/utils"
	"github.com/julienschmidt/httprouter"

	"net/http"
)

type LoginType string

const (
	LoginTypePassword LoginType = "m.login.password"
	LoginTypeEmail              = "m.login.email.identity"
)

type AuthFlow struct {
	Stages []LoginType `json:"stages,omitempty"`
	Type   LoginType   `json:"type"`
}

type AuthFlows struct {
	Flows []AuthFlow `json:"flows"`
}

type authRequest struct {
	Type     LoginType `json:"type"`
	Username string    `json:"user"`
	Password string    `json:"password"`
}

type authResponse struct {
	UserId      types.UserId `json:"user_id"`
	AccessToken string       `json:"access_token"`
}

var defaultRegisterFlows = AuthFlows{
	Flows: []AuthFlow{
		{
			Stages: []LoginType{ // not implemented
				LoginTypeEmail,
				LoginTypePassword,
			},
			Type: LoginTypeEmail,
		},
		{Type: LoginTypePassword},
	},
}

var defaultLoginFlows = AuthFlows{
	Flows: []AuthFlow{
		{Type: LoginTypePassword},
	},
}

func (e authEndpoint) registerWithPassword(hostname string, body *authRequest) interface{} {
	if body.Username == "" {
		body.Username = utils.RandomString(24)
	}
	if body.Password == "" {
		return types.BadJsonError("Missing or invalid password")
	}
	userId := types.NewUserId(body.Username, hostname)
	err := e.userService.CreateUser(userId)
	if err != nil {
		return err
	}
	if err := e.userService.SetPassword(userId, userId, body.Password); err != nil {
		return err
	}
	accessToken, err := e.tokenService.NewAccessToken(userId)
	if err != nil {
		return err
	}
	return authResponse{
		UserId:      userId,
		AccessToken: accessToken.String(),
	}
}

func (e authEndpoint) postRegister(req *http.Request, body *authRequest) interface{} {
	switch body.Type {
	case LoginTypePassword:
		hostname := strings.Split(req.Host, ":")[0]
		return e.registerWithPassword(hostname, body)
	}
	return types.BadJsonError(fmt.Sprintf("Missing or invalid login type: '%s'", body.Type))
}

func (e authEndpoint) loginWithPassword(hostname string, body *authRequest) interface{} {
	if body.Username == "" {
		return types.BadJsonError("Missing or invalid user")
	}
	if body.Password == "" {
		return types.BadJsonError("Missing or invalid password")
	}
	user := types.NewUserId(body.Username, hostname)
	err := e.userService.UserExists(user, user)
	if err != nil {
		return err
	}
	if err := e.userService.VerifyPassword(user, body.Password); err != nil {
		return err
	}
	accessToken, err := e.tokenService.NewAccessToken(user)
	if err != nil {
		return err
	}
	return authResponse{
		UserId:      user,
		AccessToken: accessToken.String(),
	}
}

func (e authEndpoint) postLogin(req *http.Request, body *authRequest) interface{} {
	switch body.Type {
	case LoginTypePassword:
		hostname := strings.Split(req.Host, ":")[0]
		return e.loginWithPassword(hostname, body)
	}
	return types.BadJsonError(fmt.Sprintf("Missing or invalid login type: '%s'", body.Type))
}

func (e authEndpoint) Register(mux *httprouter.Router) {
	mux.GET("/register", jsonHandler(func() interface{} {
		return &defaultRegisterFlows
	}))
	mux.GET("/login", jsonHandler(func() interface{} {
		return &defaultLoginFlows
	}))
	mux.POST("/register", jsonHandler(e.postRegister))
	mux.POST("/login", jsonHandler(e.postLogin))
}

type authEndpoint struct {
	userService  interfaces.UserService
	tokenService interfaces.TokenService
}

func NewAuthEndpoint(
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
) Endpoint {
	return authEndpoint{
		userService:  userService,
		tokenService: tokenService,
	}
}
