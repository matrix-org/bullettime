package service

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Rugvip/bullettime/interfaces"
	"github.com/Rugvip/bullettime/types"
	"github.com/Rugvip/bullettime/utils"
)

func CreateTokenService() (interfaces.TokenService, error) {
	return tokenService{}, nil
}

type tokenService struct{}

type tokenInfo struct {
	userId types.UserId
}

func (t tokenInfo) String() string {
	encodedUserId := base64.URLEncoding.EncodeToString([]byte(t.userId.String()))
	encodedUserId = strings.TrimRight(encodedUserId, "=")
	return fmt.Sprintf("%s..%s", encodedUserId, utils.RandomString(16))
}

func (t tokenInfo) UserId() types.UserId {
	return t.userId
}

func (t tokenService) NewAccessToken(userId types.UserId) (interfaces.Token, types.Error) {
	return tokenInfo{userId}, nil
}

func (t tokenService) ParseAccessToken(token string) (interfaces.Token, types.Error) {
	splits := strings.Split(token, "..")
	if len(splits) != 2 {
		return nil, types.DefaultUnknownTokenError
	}
	userIdStr, err := base64.URLEncoding.DecodeString(splits[0])
	if err != nil {
		return nil, types.DefaultUnknownTokenError
	}
	userId, err := types.ParseUserId(string(userIdStr))
	if err != nil {
		return nil, types.DefaultUnknownTokenError
	}
	return tokenInfo{userId}, nil
}
