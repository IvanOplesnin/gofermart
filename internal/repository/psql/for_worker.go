package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IvanOplesnin/gofermart.git/internal/logger"
	"github.com/IvanOplesnin/gofermart.git/internal/repository/psql/query"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Repo) ListPending(ctx context.Context, limit int32, statuses []string, timeSync time.Time) ([]gophermart.Order, error) {
	args := query.ListPendingParams{
		Limit:      limit,
		NextSyncAt: pgtype.Timestamptz{Valid: true, Time: timeSync},
		Statuses:   statuses,
	}
	row, err := r.queries.ListPending(ctx, args)
	if err == nil {
		orders := make([]gophermart.Order, 0, len(row))
		for _, order := range row {
			orders = append(orders, gophermart.Order{
				UserID:      order.UserID,
				Number:      order.Number,
				OrderStatus: order.OrderStatus,
				UploadedAt:  order.UploadedAt.Time,
			})
		}
		return orders, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return []gophermart.Order{}, nil
	}
	return nil, fmt.Errorf("repo.ListPending error: %w", err)
}

func (r *Repo) UpdateFromAccrual(ctx context.Context, number string, status string, nextSync time.Time) error {
	if err := r.queries.UpdateFromAccrual(ctx, query.UpdateFromAccrualParams{
		Number:     number,
		Status:     status,
		NextSyncAt: pgtype.Timestamptz{Valid: true, Time: nextSync},
	}); err != nil {
		return fmt.Errorf("repo.UpdateFromAccrual error: %w", err)
	}
	return nil
}

func (r *Repo) UpdateSyncTime(ctx context.Context, number string, nextSync time.Time) error {
	if err := r.queries.UpdateSyncTime(ctx, query.UpdateSyncTimeParams{
		Number:     number,
		NextSyncAt: pgtype.Timestamptz{Valid: true, Time: nextSync},
	}); err != nil {
		return fmt.Errorf("repo.UpdateSyncTime error: %w", err)
	}
	return nil
}

func (r *Repo) ApplyAccrual(ctx context.Context, number string, accrual int64, userID int32) error {
	err := r.InTx(ctx, func(rTx *Repo) error {
		paramsMark := query.MarkOrderProcessedParams{
			Number:  number,
			Accrual: pgtype.Int4{Valid: true, Int32: int32(accrual)},
			UserID:  int32(userID),
		}
		markRow, err := rTx.queries.MarkOrderProcessed(ctx, paramsMark)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Warn("MarkOrderProcessed: no rows")
			return nil
		}
		if err != nil {
			return fmt.Errorf("repo.ApplyAccrual: %w", err)
		}
		addBalanceParams := query.AddToUserBalanceUpsertParams{
			UserID:  markRow.UserID,
			Balance: markRow.Accrual.Int32,
		}
		if err := rTx.queries.AddToUserBalanceUpsert(ctx, addBalanceParams); err != nil {
			return fmt.Errorf("repo.ApplyAccrual: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("repo.ApplyAccrual: %w", err)
	}
	return nil
}
