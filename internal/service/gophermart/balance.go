package gophermart

import (
	"context"
	"errors"
	"fmt"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
	"github.com/jackc/pgx/v5"
)

type BalanceDb interface {
	Balance(ctx context.Context, userId int32) (Balance, error)
}

type Balance struct {
	Id          int32
	UserId      int32
	Balance     int32
	Withdraw    int32
}

func (s *Service) Balance(ctx context.Context) (handler.BalanceResponse, error) {
	const msg = "service.Orders"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	userId, err := handler.UserIdFromCtx(ctx)
	if err != nil {
		return handler.BalanceResponse{}, wrapError(err)
	}

	balance, err := s.balanceDb.Balance(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return handler.BalanceResponse{}, nil
	}
	if err != nil {
		return handler.BalanceResponse{}, wrapError(err)
	}

	return handler.BalanceResponse{
		Current: float64(balance.Balance) / 100,
		Withdrawn: float64(balance.Withdraw) / 100,
	}, nil
}
