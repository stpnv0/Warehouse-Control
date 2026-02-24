package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type ItemRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewItemRepository(db *dbpg.DB, strategy retry.Strategy) *ItemRepository {
	return &ItemRepository{
		db:       db,
		strategy: strategy,
	}
}

func (r *ItemRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.CreateItemInput,
) (*domain.Item, error) {
	const op = "ItemRepository.Create"

	query := `INSERT INTO items (name, sku, quantity, price, location)
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING id, name, sku, quantity, price, location, created_at, updated_at`

	var i domain.Item
	err := withAuditContext(ctx, r.db, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(
			ctx, query, input.Name, input.SKU, input.Quantity,
			input.Price.StringFixed(2), input.Location,
		).Scan(
			&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price,
			&i.Location, &i.CreatedAt, &i.UpdatedAt,
		)
	})

	if err != nil {
		if isDuplicateKey(err) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrDuplicateSKU)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &i, nil
}

func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	const op = "ItemRepository.GetByID"

	query := `SELECT id, name, sku, quantity, price, location, created_at, updated_at
			  FROM items
			  WHERE id=$1`

	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var i domain.Item
	if err = row.Scan(
		&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price,
		&i.Location, &i.CreatedAt, &i.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s - scan item: %w", op, err)
	}

	return &i, nil
}

func (r *ItemRepository) List(
	ctx context.Context,
	filter *domain.ItemFilter,
	limit, offset int,
) ([]*domain.Item, int64, error) {
	const op = "ItemRepository.List"

	var (
		conditions []string
		args       []interface{}
		argIdx     = 1
	)
	if filter.Search != nil && *filter.Search != "" {
		search := "%" + *filter.Search + "%"
		conditions = append(conditions,
			fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", argIdx, argIdx),
		)
		args = append(args, search)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	query := fmt.Sprintf(`
		SELECT 
		    id, name, sku, quantity, price, location, created_at, updated_at,
			COUNT(*) OVER() AS total_count
		FROM items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	listArgs := append(args, limit, offset)

	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var res []*domain.Item
	for rows.Next() {
		var i domain.Item
		if err = rows.Scan(
			&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price,
			&i.Location, &i.CreatedAt, &i.UpdatedAt, &totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("%s - scan item: %w", op, err)
		}

		res = append(res, &i)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	if res == nil {
		res = []*domain.Item{}
	}

	return res, totalCount, nil
}

func (r *ItemRepository) Update(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
	input *domain.UpdateItemInput,
) (*domain.Item, error) {
	const op = "ItemRepository.Update"

	var (
		setClauses []string
		args       []interface{}
		argIdx     = 1
	)

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.SKU != nil {
		setClauses = append(setClauses, fmt.Sprintf("sku = $%d", argIdx))
		args = append(args, *input.SKU)
		argIdx++
	}
	if input.Quantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("quantity = $%d", argIdx))
		args = append(args, *input.Quantity)
		argIdx++
	}
	if input.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, input.Price.StringFixed(2))
		argIdx++
	}
	if input.Location != nil {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *input.Location)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrNoChanges)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE items
		SET %s
		WHERE id=$%d
		RETURNING id, name, sku, quantity, price, location, created_at, updated_at
		`, strings.Join(setClauses, ", "), argIdx)

	var i domain.Item
	err := withAuditContext(ctx, r.db, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(
			&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price,
			&i.Location, &i.CreatedAt, &i.UpdatedAt,
		)
	})

	if err != nil {
		if isDuplicateKey(err) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrDuplicateSKU)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &i, nil
}

func (r *ItemRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	const op = "ItemRepository.Delete"

	query := `DELETE FROM items WHERE id=$1`

	err := withAuditContext(ctx, r.db, userID, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrNotFound
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
