package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/IvanOplesnin/gofermart.git/internal/repository/psql/query"
	"github.com/IvanOplesnin/gofermart.git/internal/service/gophermart"
	"github.com/jackc/pgx/v5/pgconn"
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
