package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type UserRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewUserRepository(db *dbpg.DB, strategy retry.Strategy) *UserRepository {
	return &UserRepository{
		db:       db,
		strategy: strategy,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	const op = "UserRepository.Create"

	query := `INSERT INTO users (username, password_hash, role) 
			  VALUES ($1, $2, $3) 
			  RETURNING id`
	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		if isDuplicateKey(err) {
			return uuid.Nil, fmt.Errorf("%s: %w", op, domain.ErrAlreadyExists)
		}
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	if err = row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("%s - scan id: %w", op, err)
	}
	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const op = "UserRepository.GetByID"

	query := `SELECT id, username, password_hash, role, created_at, updated_at
			  FROM users
			  WHERE id=$1`

	var u domain.User
	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s - scan user: %w", op, err)
	}

	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const op = "UserRepository.GetByUsername"

	query := `SELECT id, username, password_hash, role, created_at, updated_at
			  FROM users
			  WHERE username=$1`

	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var u domain.User
	if err = row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s - scan user: %w", op, err)
	}

	return &u, nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	const op = "UserRepository.List"

	query := `SELECT id, username, password_hash, role, created_at, updated_at
			  FROM users
			  ORDER BY username`

	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var res []*domain.User
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(
			&u.ID, &u.Username, &u.PasswordHash,
			&u.Role, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s - scan user: %w", op, err)
		}

		res = append(res, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}
