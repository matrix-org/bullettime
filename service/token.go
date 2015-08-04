package service

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Rugvip/bullettime/types"
	"github.com/Rugvip/bullettime/utils"
)

type TokenInfo struct {
	UserId types.UserId
}

func NewAccessToken(userId types.UserId) (string, error) {
	encodedUserId := base64.URLEncoding.EncodeToString([]byte(userId.String()))
	encodedUserId = strings.TrimRight(encodedUserId, "=")
	return fmt.Sprintf("%s..%s", encodedUserId, utils.RandomString(16)), nil
}

func ParseAccessToken(token string) (TokenInfo, error) {
	var info TokenInfo
	splits := strings.Split(token, "..")
	if len(splits) != 2 {
		return info, types.DefaultUnknownTokenError
	}
	userIdStr, err := base64.URLEncoding.DecodeString(splits[0])
	if err != nil {
		return info, err
	}
	userId, err := types.ParseUserId(string(userIdStr))
	if err != nil {
		return info, err
	}
	info.UserId = userId
	return info, nil
}
