package api

import (
	"fmt"
	"strings"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
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
		return types.BadJsonError("Missing or invalid user")
	}
	if body.Password == "" {
		return types.BadJsonError("Missing or invalid password")
	}
	userId := types.NewUserId(body.Username, hostname)
	user, err := e.userService.CreateUser(userId)
	if err != nil {
		return err
	}
	if err := user.SetPassword(body.Password); err != nil {
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

func (e authEndpoint) postRegister(req *http.Request, params httprouter.Params, body *authRequest) interface{} {
	switch body.Type {
	case LoginTypePassword:
		hostname := strings.Split(req.Host, ":")[0]
		return e.registerWithPassword(hostname, body)
	}
	return types.BadJsonError(fmt.Sprintf("Missing or invalid login type: '%s'", body.Type))
}

func (e authEndpoint) postLogin() interface{} {
	return "login"
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
