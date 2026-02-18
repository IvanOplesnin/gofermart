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

func (r *Repo) AddUser(ctx context.Context, login string, passwordHash string) (uint64, error) {
	argAddUser := query.AddUserParams{
		Login:        login,
		PasswordHash: passwordHash,
	}
	userID, err := r.queries.AddUser(ctx, argAddUser)
	if err == nil {
		return uint64(userID), nil
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
		ID:           uint64(dbUser.ID),
		Login:        dbUser.Login,
		HashPassword: dbUser.PasswordHash,
	}, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id uint64) (uint64, error) {
	userId, err := r.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, gophermart.ErrNoRow
	}
	if err != nil {
		return 0, fmt.Errorf("repo.GetUserById error: %w", err)
	}
	return uint64(userId), nil
}


func (r *Repo) CreateOrder(ctx context.Context, userID uint64, number string) (bool, uint64, error) {
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
		return false, uint64(order.UserID), nil
	}
	return false, 0, fmt.Errorf("repo.CreateOrder:: %w", err)
}


func (r *Repo) GetOrders(ctx context.Context, userId int32) ([]gophermart.Order, error) {
	return []gophermart.Order{}, nil
}
