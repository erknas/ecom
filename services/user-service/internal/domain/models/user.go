package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	FirstName    string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}

type UpdatedUser struct {
	FirstName    *string
	Email        *string
	PasswordHash []byte
}

func NewUser(firstName, email, password string) (*User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		FirstName:    firstName,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}, nil
}

func NewUpdatedUser(firstName, email, password *string) (*UpdatedUser, error) {
	if password == nil {
		return &UpdatedUser{
			FirstName:    firstName,
			Email:        email,
			PasswordHash: nil,
		}, nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &UpdatedUser{
		FirstName:    firstName,
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}
