package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/prudhvinik1/edgesync/internal/models"
)

type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	GetByEmail(ctx context.Context, email string) (*models.Account, error)
	Update(ctx context.Context, account *models.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DeviceRepository interface {
	Create(ctx context.Context, device *models.Device) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error)
	GetDevicesByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Device, error)
	Update(ctx context.Context, device *models.Device) error
	Revoke(ctx context.Context, id uuid.UUID) error
}

type EncryptedStateRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.EncryptedState, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.EncryptedState, error)
	GetByKey(ctx context.Context, accountID uuid.UUID, key string) (*models.EncryptedState, error)
	Upsert(ctx context.Context, state *models.EncryptedState) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForAccount(ctx context.Context, accountID uuid.UUID) error
}

type SyncEventRepository interface {
	Append(ctx context.Context, event *models.SyncEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SyncEvent, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.SyncEvent, error)
	GetSinceSequence(ctx context.Context, accountID uuid.UUID, sequenceNumber int64) ([]*models.SyncEvent, error)
}

type PresenceRepository interface {
	SetPresence(ctx context.Context, presence *models.Presence) error
	GetPresence(ctx context.Context, deviceID uuid.UUID) (*models.Presence, error)
	DeletePresence(ctx context.Context, deviceID uuid.UUID) error
	GetBulkPresence(ctx context.Context, deviceIDs []uuid.UUID) (map[uuid.UUID]models.Presence, error)
}
