package db

import "github.com/Rugvip/bullettime/types"

type dbUser struct {
	types.UserId
	types.UserProfile
	PasswordHash string `json:"-"`
}

var userTable = map[types.UserId]*dbUser{}

func CreateUser(id types.UserId) types.Error {
	if userTable[id] != nil {
		return types.UserInUseError("user '" + id.String() + "' already exists")
	}
	user := new(dbUser)
	user.UserId = id
	userTable[id] = user
	return nil
}

func UserExists(id types.UserId) types.Error {
	if userTable[id] == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return nil
}

func SetUserPasswordHash(id types.UserId, hash string) types.Error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.PasswordHash = hash
	return nil
}

func UserPasswordHash(id types.UserId) (string, types.Error) {
	user := userTable[id]
	if user == nil {
		return "", types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.PasswordHash, nil
}

func SetUserDisplayName(id types.UserId, displayName string) types.Error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.DisplayName = displayName
	return nil
}

func SetUserAvatarUrl(id types.UserId, avatarUrl string) types.Error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.AvatarUrl = avatarUrl
	return nil
}

func UserProfile(id types.UserId) (types.UserProfile, types.Error) {
	user := userTable[id]
	if user == nil {
		return types.UserProfile{}, types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.UserProfile, nil
}
