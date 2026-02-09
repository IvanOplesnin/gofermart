package gophermart

import (
	"context"
	"errors"
)


var (
	ErrUserAlreadyExists = errors.New("user already exist")
)

type UserAdd interface {
	AddUser(ctx context.Context, login string, password_hash string) (uint64, error)
}


type Hasher interface {
	HashPassword(password string) (string, error)
	ComparePasswordHash(password string, hash string) (bool, error)
}