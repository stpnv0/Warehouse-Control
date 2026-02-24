package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
)

func isDuplicateKey(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// withAuditContext выполняет fn внутри транзакции с установленным app.current_user_id (необходимо для триггера аудита)
func withAuditContext(ctx context.Context, db *dbpg.DB, userID uuid.UUID, fn func(tx *sql.Tx) error) error {
	return db.WithTx(ctx, func(tx *sql.Tx) error {
		queryAudit := `SELECT set_config('app.current_user_id', $1, true)`
		_, err := tx.ExecContext(ctx, queryAudit, userID.String())
		if err != nil {
			return fmt.Errorf("set audit context: %w", err)
		}

		return fn(tx)
	})
}
