package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IvanOplesnin/gofermart.git/internal/repository/psql/query"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type Repo struct {
	db      *pgxpool.Pool
	queries *query.Queries
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{
		db:      db,
		queries: query.New(db),
	}
}

func (r *Repo) AddUser(ctx context.Context, login string, passwordHash string) (int32, error) {
	argAddUser := query.AddUserParams{
		Login:        login,
		PasswordHash: passwordHash,
	}
	userID, err := r.queries.AddUser(ctx, argAddUser)
	if err == nil {
		return userID, nil
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == pgUniqueViolation {
		return 0, gophermart.ErrUserAlreadyExists
	}
	return 0, fmt.Errorf("repo.AddUser error: %w", err)
}

func (r *Repo) GetUserByLogin(ctx context.Context, login string) (gophermart.User, error) {
	dbUser, err := r.queries.GetUserByLogin(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return gophermart.User{}, gophermart.ErrNoRow
	}
	if err != nil {
		return gophermart.User{}, fmt.Errorf("repo.GetUserByLogin error: %w", err)
	}
	return gophermart.User{
		ID:           dbUser.ID,
		Login:        dbUser.Login,
		HashPassword: dbUser.PasswordHash,
	}, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id int32) (int32, error) {
	userId, err := r.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, gophermart.ErrNoRow
	}
	if err != nil {
		return 0, fmt.Errorf("repo.GetUserById error: %w", err)
	}
	return userId, nil
}

func (r *Repo) CreateOrder(ctx context.Context, userID int32, number string) (bool, int32, error) {
	argCreateOrder := query.AddOrderParams{
		UserID:     int32(userID),
		Number:     number,
		Status:     "NEW",
		UploadedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	err := r.queries.AddOrder(ctx, argCreateOrder)
	if err == nil {
		return true, 0, nil
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == pgUniqueViolation {
		order, err := r.queries.GetOrderByNumber(ctx, number)
		if err != nil {
			return false, 0, fmt.Errorf("repo.CreateOrder: %w", err)
		}
		return false, order.UserID, nil
	}
	return false, 0, fmt.Errorf("repo.CreateOrder:: %w", err)
}

func (r *Repo) GetOrders(ctx context.Context, userId int32) ([]gophermart.Order, error) {
	rows, err := r.queries.GetOrdersByUserID(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return []gophermart.Order{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repo.GetOrders error: %w", err)
	}
	orders := make([]gophermart.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, gophermart.Order{
			UserID:      row.UserID,
			Number:      row.Number,
			OrderStatus: row.Status,
			Accrual:     row.Accrual.Int32,
			UploadedAt:  row.UploadedAt.Time,
		})
	}
	return orders, nil
}

func (r *Repo) ListWithdraws(ctx context.Context, userId int32) ([]gophermart.Withdraw, error) {
	var result []gophermart.Withdraw
	err := r.InTx(ctx, func(rTx *Repo) error {
		withdraws, err := rTx.queries.ListWithdraws(ctx, userId)
		if err != nil {
			return err
		}
		if len(withdraws) == 0 {
			result = []gophermart.Withdraw{}
			return gophermart.ErrNoRow
		}
		result = make([]gophermart.Withdraw, 0, len(withdraws))
		for _, w := range withdraws {
			result = append(result, gophermart.Withdraw{
				ID:          w.ID,
				UserID:      w.UserID,
				OrderNumber: w.OrderNumber,
				Summa:       w.Summa,
				ProcessedAt: w.ProcessedAt.Time,
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repo.ListWithdraws: %w", err)
	}
	return result, nil
}

func (r *Repo) Withdraw(ctx context.Context, userId int32, summa int32, order string) error {
	err := r.InTx(ctx, func(rTx *Repo) error {
		if err := rTx.queries.EnsureBalanceRow(ctx, userId); err != nil {
			return err
		}
		_, err := rTx.queries.WithdrawIfEnough(ctx, query.WithdrawIfEnoughParams{
			UserID: userId,
			Summa:  summa,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return gophermart.ErrNotEnoughBalance
		}
		if err != nil {
			return err
		}
		err = rTx.queries.AddWithdrawal(ctx, query.AddWithdrawalParams{
			UserID:      userId,
			OrderNumber: order,
			Summa:       summa,
			ProcessedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				return gophermart.ErrWithdrawAlreadyProcessed
			}
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("repo.Withdraw: %w", err)
	}
	return nil
}


func (r *Repo) Balance(ctx context.Context, userId int32) (gophermart.Balance, error) {
	var balance gophermart.Balance
	err := r.InTx(ctx, func(rTx *Repo) error {
		err := rTx.queries.EnsureBalanceRow(ctx, userId)
		if err != nil {
			return err
		}
		balanceRow, err := rTx.queries.BalnceByUserID(ctx, userId)
		if errors.Is(err, pgx.ErrNoRows) {
			return gophermart.ErrNoRow
		}
		if err != nil {
			return err
		}
		balance = gophermart.Balance{
			Id:       balanceRow.ID,
			UserId:   balanceRow.UserID,
			Balance:  balanceRow.Balance,
			Withdraw: balanceRow.Withdrawn,
		}
		return nil
	})
	if err != nil {
		return gophermart.Balance{}, fmt.Errorf("repo.Balance: %w", err)
	}
	return balance, nil
}
