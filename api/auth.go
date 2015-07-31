package api

import "net/http"

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

type loginBody struct {
	Type     LoginType `json:"type"`
	Username string    `json:"user"`
	Password string    `json:"password"`
}

type authResponse struct {
	UserId      string `json:"user_id"`
	AccessToken string `json:"access_token"`
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

var registerResource = Resource{
	Get: NewJsonHandler(func(req *http.Request) interface{} {
		return &defaultRegisterFlows
	}),
}

var loginResource = Resource{
	Get: NewJsonHandler(func(req *http.Request) interface{} {
		return &defaultLoginFlows
	}),
}
