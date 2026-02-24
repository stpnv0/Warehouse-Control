package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type AuditRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewAuditRepository(db *dbpg.DB, strategy retry.Strategy) *AuditRepository {
	return &AuditRepository{
		db:       db,
		strategy: strategy,
	}
}

func (r *AuditRepository) GetByItemID(ctx context.Context, itemID uuid.UUID) ([]*domain.AuditEntryWithUser, error) {
	const op = "AuditRepository.GetByItemID"

	query := `
		SELECT
			a.id, a.item_id, a.action, a.changed_by,
			a.old_data, a.new_data, a.diff, a.changed_at,
			COALESCE(u.username, 'unknown') AS username
		FROM item_audit_log a
		LEFT JOIN users u ON u.id = a.changed_by
		WHERE item_id=$1
		ORDER BY a.changed_at DESC`

	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var res []*domain.AuditEntryWithUser
	for rows.Next() {
		e, err := scanAuditRow(rows)
		if err != nil {
			return nil, fmt.Errorf("%s - scan audit: %w", op, err)
		}
		res = append(res, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (r *AuditRepository) List(
	ctx context.Context,
	filter *domain.AuditFilter,
	limit, offset int,
) ([]*domain.AuditEntryWithUser, int64, error) {
	const op = "AuditRepository.List"

	var (
		conditions []string
		args       []interface{}
		argIdx     = 1
	)
	if filter.ItemID != nil {
		conditions = append(conditions, fmt.Sprintf("a.item_id = $%d", argIdx))
		args = append(args, *filter.ItemID)
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("a.changed_by = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argIdx))
		args = append(args, string(*filter.Action))
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("a.changed_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("a.changed_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT 
			a.id, a.item_id, a.action, a.changed_by,
			a.old_data, a.new_data, a.diff, a.changed_at,
			COALESCE(u.username, 'unknown') AS username,
			COUNT(*) OVER() AS total_count
		FROM item_audit_log a
		LEFT JOIN users u ON u.id = a.changed_by
		%s
		ORDER BY a.changed_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	listArgs := append(args, limit, offset)
	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var (
		res        []*domain.AuditEntryWithUser
		totalCount int64
	)
	for rows.Next() {
		e, err := scanAuditRowWithTotal(rows, &totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("%s - scan audit: %w", op, err)
		}
		res = append(res, e)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return res, totalCount, nil
}

func scanAuditRow(rows *sql.Rows) (*domain.AuditEntryWithUser, error) {
	var (
		e       domain.AuditEntryWithUser
		oldData []byte
		newData []byte
		diff    []byte
	)

	if err := rows.Scan(
		&e.ID, &e.ItemID, &e.Action, &e.ChangedBy, &oldData,
		&newData, &diff, &e.ChangedAt, &e.Username,
	); err != nil {
		return nil, err
	}

	e.OldData = oldData
	e.NewData = newData
	e.Diff = diff

	return &e, nil
}

func scanAuditRowWithTotal(rows *sql.Rows, totalCount *int64) (*domain.AuditEntryWithUser, error) {
	var (
		e       domain.AuditEntryWithUser
		oldData []byte
		newData []byte
		diff    []byte
	)

	if err := rows.Scan(
		&e.ID, &e.ItemID, &e.Action, &e.ChangedBy,
		&oldData, &newData, &diff, &e.ChangedAt,
		&e.Username,
		totalCount,
	); err != nil {
		return nil, err
	}

	e.OldData = oldData
	e.NewData = newData
	e.Diff = diff
	return &e, nil
}
