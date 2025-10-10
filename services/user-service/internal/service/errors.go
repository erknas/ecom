package service

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrGeneratePasswordHash = errors.New("generate password hash")
)
