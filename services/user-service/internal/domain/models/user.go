package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	FirstName    string
	PhoneNumber  string
	Email        string
	PasswordHash []byte
	CreatedAt    time.Time
}

func NewUser(firstName, phoneNumber, email, password string) (*User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		FirstName:    firstName,
		PhoneNumber:  phoneNumber,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}, nil
}
