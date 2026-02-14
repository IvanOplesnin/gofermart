package gophermart

import (
	"context"
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user already exist")
)

type UserCRUD interface {
	AddUser(ctx context.Context, login string, password_hash string) (uint64, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
	GetUserByID(ctx context.Context, id uint64) (uint64, error)
}

type User struct {
	ID           uint64 `json:"id"`
	Login        string `json:"login"`
	HashPassword string `json:"hash_password"`
}

type UserGetter interface {
}

type Hasher interface {
	HashPassword(password string) (string, error)
	ComparePasswordHash(password string, hash string) (bool, error)
}

type GetApiOrdered interface {
	GetOrder(ctx context.Context, number string) (status int, accrual float64, err error)
}

type AddOrdered interface {
	CreateOrder(ctx context.Context, userID uint64, number string) (created bool, ownerUserID uint64, err error)
}
