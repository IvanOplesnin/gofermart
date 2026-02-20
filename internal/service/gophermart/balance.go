package gophermart

import (
	"context"
	"errors"
	"fmt"

	"github.com/IvanOplesnin/gofermart.git/internal/handler"
	"github.com/jackc/pgx/v5"
)

type BalanceDB interface {
	Balance(ctx context.Context, userID int32) (Balance, error)
}

type Balance struct {
	ID       int32
	UserID   int32
	Balance  int32
	Withdraw int32
}

func (s *Service) Balance(ctx context.Context) (handler.BalanceResponse, error) {
	const msg = "service.Orders"
	wrapError := func(err error) error { return fmt.Errorf("%s: %w", msg, err) }

	userID, err := handler.UserIDFromCtx(ctx)
	if err != nil {
		return handler.BalanceResponse{}, wrapError(err)
	}

	balance, err := s.balanceDB.Balance(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return handler.BalanceResponse{}, nil
	}
	if err != nil {
		return handler.BalanceResponse{}, wrapError(err)
	}

	return handler.BalanceResponse{
		Current:   float64(balance.Balance) / 100,
		Withdrawn: float64(balance.Withdraw) / 100,
	}, nil
}
