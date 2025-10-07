package storage

import "errors"

var (
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrNoChanges        = errors.New("nothing to update")
	ErrInternalDatabase = errors.New("internal database")
)
