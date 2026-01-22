package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prudhvinik1/edgesync/internal/models"
)

var ErrNotFound = errors.New("not found")

type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAccountRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool}
}

func (r *PostgresAccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `INSERT INTO accounts (email, password_hash) 
              VALUES ($1, $2) 
              RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, account.Email, account.PasswordHash).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func (r *PostgresAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at, deleted_at FROM accounts WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var account models.Account
	err := row.Scan(&account.ID, &account.Email, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt, &account.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &account, nil
}

func (r *PostgresAccountRepository) GetByEmail(ctx context.Context, email string) (*models.Account, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at, deleted_at FROM accounts WHERE email = $1`

	row := r.pool.QueryRow(ctx, query, email)

	var account models.Account
	err := row.Scan(&account.ID, &account.Email, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt, &account.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &account, nil
}

func (r *PostgresAccountRepository) Update(ctx context.Context, account *models.Account) error {
	query := `UPDATE accounts SET email = $1, password_hash = $2, updated_at = NOW() WHERE id = $3`

	result, err := r.pool.Exec(ctx, query, account.Email, account.PasswordHash, account.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE accounts SET deleted_at = NOW() WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
