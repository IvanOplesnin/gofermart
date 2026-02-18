package gophermart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
)

type WithdrawerDb interface {
	Withdraw(ctx context.Context, userId int32, summa int32, order string) error
	ListWithdraws(ctx context.Context, userId int32) ([]Withdraw, error)
}

type Withdraw struct {
	ID          int32
	UserID      int32
	OrderNumber string
	Summa       int32
	ProcessedAt time.Time
}

var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrWithdrawAlreadyProcessed = errors.New("withdraw already processed")

func (s *Service) Withdraw(ctx context.Context, orderNumber string, summa float64) error {
	const msg = "service.Orders"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	if summa <= 0 {
		return handler.ErrInvalidSumma
	}
	userId, err := handler.UserIdFromCtx(ctx)
	if err != nil {
		return wrapError(err)
	}
	err = s.withdrawDb.Withdraw(ctx, int32(userId), int32(summa*100), orderNumber)
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotEnoughBalance) {
		return handler.ErrEnoughMoney
	}
	if errors.Is(err, ErrWithdrawAlreadyProcessed) {
		return handler.ErrInvalidOrderNumber
	}
	return wrapError(err)
}
