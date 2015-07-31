package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Rugvip/bullettime/utils"
)

type TokenInfo struct {
	userId string
}

func NewAccessToken(userId string) (string, error) {
	encodedUserId := base64.URLEncoding.EncodeToString([]byte(userId))
	encodedUserId = strings.TrimRight(encodedUserId, "=")
	return fmt.Sprintf("%s..%s", encodedUserId, utils.RandomString(16)), nil
}

func ParseAccessToken(token string) (TokenInfo, error) {
	var info TokenInfo
	splits := strings.Split(token, "..")
	if len(splits) != 2 {
		return info, errors.New("failed to parse token")
	}
	userId, err := base64.URLEncoding.DecodeString(splits[0])
	if err != nil {
		return info, err
	}
	info.userId = string(userId)
	return info, nil
}
