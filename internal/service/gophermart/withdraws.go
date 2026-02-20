package gophermart

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
)

type WithdrawerDB interface {
	Withdraw(ctx context.Context, userID int32, summa int32, order string) error
	ListWithdraws(ctx context.Context, userID int32) ([]Withdraw, error)
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
	userID, err := handler.UserIDFromCtx(ctx)
	if err != nil {
		return wrapError(err)
	}
	err = s.withdrawDB.Withdraw(ctx, int32(userID), int32(math.Round(summa*100)), orderNumber)
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

func (s *Service) ListWithdraws(ctx context.Context) ([]handler.Withdraw, error) {
	const msg = "service.ListWithdraws"
	wrapErr := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	userID, err := handler.UserIDFromCtx(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	withdraws, err := s.withdrawDB.ListWithdraws(ctx, int32(userID))
	if err != nil {
		return nil, wrapErr(err)
	}
	if len(withdraws) == 0 {
		return []handler.Withdraw{}, nil
	}
	respWithdraws := make([]handler.Withdraw, 0, len(withdraws))
	for _, w := range withdraws {
		respWithdraws = append(respWithdraws, handler.Withdraw{
			OrderNumber: w.OrderNumber,
			Summa:       float64(w.Summa) / 100,
			ProcessedAt: handler.RFC3339Time(w.ProcessedAt),
		})
	}
	return respWithdraws, nil
}
