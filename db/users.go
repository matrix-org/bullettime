package db

import "github.com/Rugvip/bullettime/types"

type User struct {
	types.UserId
	types.UserProfile
	PasswordHash string `json:"-"`
}

var userTable = map[types.UserId]*User{}

func CreateUser(id types.UserId) error {
	if userTable[id] != nil {
		return types.UserInUseError("user '" + id.String() + "' already exists")
	}
	user := new(User)
	user.UserId = id
	userTable[id] = user
	return nil
}

func UserExists(id types.UserId) error {
	if userTable[id] == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return nil
}

func SetUserPasswordHash(id types.UserId, hash string) error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.PasswordHash = hash
	return nil
}

func GetUserPasswordHash(id types.UserId) (string, error) {
	user := userTable[id]
	if user == nil {
		return "", types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.PasswordHash, nil
}

func SetUserDisplayName(id types.UserId, displayName string) error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.DisplayName = displayName
	return nil
}

func SetUserAvatarUrl(id types.UserId, avatarUrl string) error {
	user := userTable[id]
	if user == nil {
		return types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	user.AvatarUrl = avatarUrl
	return nil
}

func GetUserProfile(id types.UserId) (types.UserProfile, error) {
	user := userTable[id]
	if user == nil {
		return types.UserProfile{}, types.NotFoundError("user '" + id.String() + "' doesn't exist")
	}
	return user.UserProfile, nil
}
