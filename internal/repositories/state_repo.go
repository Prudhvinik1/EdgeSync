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

// ErrVersionConflict is returned when optimistic locking fails
var ErrVersionConflict = errors.New("version conflict: state was modified by another device")

type PostgresEncryptedStateRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEncryptedStateRepository(pool *pgxpool.Pool) *PostgresEncryptedStateRepository {
	return &PostgresEncryptedStateRepository{pool: pool}
}

func (r *PostgresEncryptedStateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EncryptedState, error) {
	query := `SELECT id, account_id, device_id, key, state, nonce, version, created_at, updated_at, deleted_at
	          FROM encrypted_states 
	          WHERE id = $1 AND deleted_at IS NULL`

	var state models.EncryptedState
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&state.ID,
		&state.AccountID,
		&state.DeviceID,
		&state.Key,
		&state.State,
		&state.Nonce,
		&state.Version,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get state by ID: %w", err)
	}
	return &state, nil
}

func (r *PostgresEncryptedStateRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.EncryptedState, error) {
	query := `SELECT id, account_id, device_id, key, state, nonce, version, created_at, updated_at, deleted_at
	          FROM encrypted_states 
	          WHERE account_id = $1 AND deleted_at IS NULL
	          ORDER BY key ASC`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query states: %w", err)
	}
	defer rows.Close()

	var states []*models.EncryptedState
	for rows.Next() {
		var state models.EncryptedState
		err := rows.Scan(
			&state.ID,
			&state.AccountID,
			&state.DeviceID,
			&state.Key,
			&state.State,
			&state.Nonce,
			&state.Version,
			&state.CreatedAt,
			&state.UpdatedAt,
			&state.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		states = append(states, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating states: %w", err)
	}

	return states, nil
}

func (r *PostgresEncryptedStateRepository) GetByKey(ctx context.Context, accountID uuid.UUID, key string) (*models.EncryptedState, error) {
	query := `SELECT id, account_id, device_id, key, state, nonce, version, created_at, updated_at, deleted_at
	          FROM encrypted_states 
	          WHERE account_id = $1 AND key = $2 AND deleted_at IS NULL`

	var state models.EncryptedState
	err := r.pool.QueryRow(ctx, query, accountID, key).Scan(
		&state.ID,
		&state.AccountID,
		&state.DeviceID,
		&state.Key,
		&state.State,
		&state.Nonce,
		&state.Version,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get state by key: %w", err)
	}
	return &state, nil
}

// Upsert creates or updates an encrypted state with optimistic locking.
// If the state doesn't exist, it creates it with version 1.
// If it exists, it only updates if the provided version matches the current version.
// On success, the state.Version is incremented and state.ID/timestamps are populated.
func (r *PostgresEncryptedStateRepository) Upsert(ctx context.Context, state *models.EncryptedState) error {
	// First, try to get existing state
	existing, err := r.GetByKey(ctx, state.AccountID, state.Key)

	if errors.Is(err, ErrNotFound) {
		// State doesn't exist - INSERT new state
		return r.create(ctx, state)
	}
	if err != nil {
		return fmt.Errorf("failed to check existing state: %w", err)
	}

	// State exists - UPDATE with optimistic locking
	return r.update(ctx, state, existing.ID)
}

// create inserts a new encrypted state
func (r *PostgresEncryptedStateRepository) create(ctx context.Context, state *models.EncryptedState) error {
	query := `INSERT INTO encrypted_states (account_id, device_id, key, state, nonce, version)
	          VALUES ($1, $2, $3, $4, $5, 1)
	          RETURNING id, version, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		state.AccountID,
		state.DeviceID,
		state.Key,
		state.State,
		state.Nonce,
	).Scan(&state.ID, &state.Version, &state.CreatedAt, &state.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create state: %w", err)
	}
	return nil
}

// update updates an existing encrypted state with optimistic locking
func (r *PostgresEncryptedStateRepository) update(ctx context.Context, state *models.EncryptedState, existingID uuid.UUID) error {
	// CRITICAL: The WHERE clause includes version check for optimistic locking
	// Only updates if the current version matches what the client expects
	query := `UPDATE encrypted_states 
	          SET device_id = $1, 
	              state = $2, 
	              nonce = $3, 
	              version = version + 1, 
	              updated_at = NOW()
	          WHERE id = $4 AND version = $5 AND deleted_at IS NULL
	          RETURNING version, updated_at`

	var newVersion int64
	err := r.pool.QueryRow(ctx, query,
		state.DeviceID,
		state.State,
		state.Nonce,
		existingID,
		state.Version, // Expected version - must match!
	).Scan(&newVersion, &state.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		// No rows updated = version mismatch = conflict!
		return ErrVersionConflict
	}
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Update the state object with new values
	state.ID = existingID
	state.Version = newVersion
	return nil
}

func (r *PostgresEncryptedStateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE encrypted_states 
	          SET deleted_at = NOW() 
	          WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
