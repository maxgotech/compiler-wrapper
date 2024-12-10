package storage

import "errors"

var (
	ErrResNotFound  = errors.New("result not found")
	ErrUserNotFound = errors.New("user not found")
)
