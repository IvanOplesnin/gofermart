package gophermart

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user already exist")
)

type UserCRUD interface {
	AddUser(ctx context.Context, login string, passwordHash string) (uint64, error)
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

type GetAPIOrdered interface {
	GetOrder(ctx context.Context, number string) (response *AccrualResponse, err error)
}

type AccrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}

func (a AccrualResponse) String() string {
	return fmt.Sprintf("{OrderNumber: %s, Status: %s, Accrual: %v}", a.OrderNumber, a.Status, a.Accrual)
}

type AddOrdered interface {
	CreateOrder(ctx context.Context, userID uint64, number string) (created bool, ownerUserID uint64, err error)
}

type ListUpdateApplyAccrual interface {
	ListPending(ctx context.Context, limit int32, statuses []string, timeSync time.Time) ([]Order, error)
	UpdateFromAccrual(ctx context.Context, number string, status string, nextSync time.Time) error
	UpdateSyncTime(ctx context.Context, number string, nextSync time.Time) error
	ApplyAccrual(ctx context.Context, number string, accrual int64, userID uint64) error
}

type Order struct {
	UserID      uint64
	Number      string
	OrderStatus string
	UploadedAt  time.Time
}
