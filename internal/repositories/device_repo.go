package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prudhvinik1/edgesync/internal/models"
)

type PostgresDeviceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDeviceRepository(pool *pgxpool.Pool) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{pool: pool}
}

func (r *PostgresDeviceRepository) Create(ctx context.Context, device *models.Device) error {
	query := `INSERT INTO devices (account_id, name, device_type, public_key) 
	          VALUES ($1, $2, $3, $4) 
	          RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		device.AccountID,
		device.Name,
		device.DeviceType,
		device.PublicKey,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}
	return nil
}

func (r *PostgresDeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	query := `SELECT id, account_id, name, device_type, public_key, 
	                 last_seen_at, revoked_at, created_at, updated_at, deleted_at 
	          FROM devices 
	          WHERE id = $1 AND deleted_at IS NULL`

	var device models.Device
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&device.ID,
		&device.AccountID,
		&device.Name,
		&device.DeviceType,
		&device.PublicKey,
		&device.LastSeenAt,
		&device.RevokedAt,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return &device, nil
}

func (r *PostgresDeviceRepository) GetDevicesByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Device, error) {
	query := `SELECT id, account_id, name, device_type, public_key, 
	                 last_seen_at, revoked_at, created_at, updated_at, deleted_at 
	          FROM devices 
	          WHERE account_id = $1 AND deleted_at IS NULL
	          ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var device models.Device
		err := rows.Scan(
			&device.ID,
			&device.AccountID,
			&device.Name,
			&device.DeviceType,
			&device.PublicKey,
			&device.LastSeenAt,
			&device.RevokedAt,
			&device.CreatedAt,
			&device.UpdatedAt,
			&device.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating devices: %w", err)
	}

	return devices, nil
}

func (r *PostgresDeviceRepository) Update(ctx context.Context, device *models.Device) error {
	query := `UPDATE devices 
	          SET name = $1, device_type = $2, public_key = $3, updated_at = NOW() 
	          WHERE id = $4 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		device.Name,
		device.DeviceType,
		device.PublicKey,
		device.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresDeviceRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE devices 
	          SET revoked_at = $1, updated_at = NOW() 
	          WHERE id = $2 AND revoked_at IS NULL AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
