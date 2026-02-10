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

type User struct {
	ID           uint64 `json:"id"`
	Login        string `json:"login"`
	HashPassword string `json:"hash_password"`
}

type UserGetter interface {
	GetUser(ctx context.Context, login string) (User, error)
}

type Hasher interface {
	HashPassword(password string) (string, error)
	ComparePasswordHash(password string, hash string) (bool, error)
}
