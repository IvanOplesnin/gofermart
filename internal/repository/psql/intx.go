package psql

import (
	"context"
	"fmt"
)

func (r *Repo) InTx(ctx context.Context, f func(r *Repo) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	txRepo := &Repo{
		db: r.db,
		queries: r.queries.WithTx(tx),
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
