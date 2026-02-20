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
	AddUser(ctx context.Context, login string, passwordHash string) (int32, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
	GetUserByID(ctx context.Context, id int32) (int32, error)
}

type User struct {
	ID           int32 `json:"id"`
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

type Ordered interface {
	CreateOrder(ctx context.Context, userID int32, number string) (created bool, ownerUserID int32, err error)
	GetOrders(ctx context.Context, userID int32) ([]Order, error)
}

type ListUpdateApplyAccrual interface {
	ListPending(ctx context.Context, limit int32, statuses []string, timeSync time.Time) ([]Order, error)
	UpdateFromAccrual(ctx context.Context, number string, status string, nextSync time.Time) error
	UpdateSyncTime(ctx context.Context, number string, nextSync time.Time) error
	ApplyAccrual(ctx context.Context, number string, accrual int64, userID int32) error
}

type Order struct {
	UserID      int32
	Number      string
	OrderStatus string
	Accrual     int32
	UploadedAt  time.Time
}
