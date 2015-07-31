package service

import (
	"github.com/Rugvip/bullettime/db"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id string
}

func GetUser(id string) (User, error) {
	if err := db.UserExists(id); err != nil {
		return User{}, err
	}
	return User{id: id}, nil
}

func CreateUser(id string) (User, error) {
	if err := db.CreateUser(id); err != nil {
		return User{}, err
	}
	return User{id: id}, nil
}

func (u User) Id() string {
	return u.id
}

func (u User) VerifyPassword(password string) error {
	hash, err := db.GetUserPasswordHash(u.id)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (u User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	if err := db.SetUserPasswordHash(u.id, string(hash)); err != nil {
		return err
	}
	return nil
}