package api

import (
	"net/http"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"

	"github.com/julienschmidt/httprouter"
)

type statusRequest struct {
	Presence      *types.Presence `json:"presence"`
	StatusMessage *string         `json:"status_msg"`
}

func (e presenceEndpoint) getStatus(req *http.Request, params httprouter.Params) interface{} {
	authedUser, err := readAccessToken(e.users, e.tokens, req)
	if err != nil {
		return err
	}
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	status, err := e.presences.Status(user, authedUser)
	if err != nil {
		return err
	}
	return status
}

func (e presenceEndpoint) setStatus(req *http.Request, params httprouter.Params, body *statusRequest) interface{} {
	authedUser, err := readAccessToken(e.users, e.tokens, req)
	if err != nil {
		return err
	}
	user, err := e.userFromParams(params)
	if err != nil {
		return err
	}
	if body.Presence == nil && body.StatusMessage == nil {
		return types.BadJsonError("empty request")
	}
	_, err = e.presences.UpdateStatus(user, authedUser, body.Presence, body.StatusMessage)
	if err != nil {
		return err
	}
	return struct{}{}
}

func (e presenceEndpoint) userFromParams(params httprouter.Params) (types.UserId, types.Error) {
	user, err := types.ParseUserId(params[0].Value)
	if err != nil {
		return types.UserId{}, types.BadJsonError(err.Error())
	}
	if err := e.users.UserExists(user, user); err != nil {
		return types.UserId{}, err
	}
	return user, nil
}

func (e presenceEndpoint) Register(mux *httprouter.Router) {
	mux.GET("/presence/:userId/status", jsonHandler(e.getStatus))
	mux.PUT("/presence/:userId/status", jsonHandler(e.setStatus))
}

type presenceEndpoint struct {
	users     interfaces.UserService
	tokens    interfaces.TokenService
	presences interfaces.PresenceService
}

func NewPresenceEndpoint(
	users interfaces.UserService,
	tokens interfaces.TokenService,
	presences interfaces.PresenceService,
) Endpoint {
	return presenceEndpoint{
		users,
		tokens,
		presences,
	}
}
