package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prudhvinik1/edgesync/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateRepository_Upsert_Create tests creating a new state (doesn't exist yet)
func TestStateRepository_Upsert_Create(t *testing.T) {
	// ARRANGE: Setup test database connection
	pool := getTestPool(t)
	repo := NewPostgresEncryptedStateRepository(pool)
	accountRepo := NewPostgresAccountRepository(pool)
	deviceRepo := NewPostgresDeviceRepository(pool)
	ctx := context.Background()

	// Create test account and device (required for foreign keys)
	accountID, deviceID := setupTestAccountAndDevice(t, ctx, pool, accountRepo, deviceRepo)
	defer cleanupTestData(t, pool, ctx, accountID)

	// ACT: Create a new state
	state := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID,
		Key:       "test-settings",
		State:     []byte("encrypted-data"),
		Nonce:     []byte("nonce-123"),
		Version:   0, // Version 0 means "doesn't exist yet"
	}

	err := repo.Upsert(ctx, state)

	// ASSERT: Should succeed and populate ID/version
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, state.ID, "ID should be generated")
	assert.Equal(t, int64(1), state.Version, "New state should start at version 1")
	assert.False(t, state.CreatedAt.IsZero(), "CreatedAt should be set")
}

// TestStateRepository_Upsert_Update tests updating an existing state successfully
func TestStateRepository_Upsert_Update(t *testing.T) {
	pool := getTestPool(t)
	repo := NewPostgresEncryptedStateRepository(pool)
	accountRepo := NewPostgresAccountRepository(pool)
	deviceRepo := NewPostgresDeviceRepository(pool)
	ctx := context.Background()

	// Create test account and device
	accountID, deviceID := setupTestAccountAndDevice(t, ctx, pool, accountRepo, deviceRepo)
	defer cleanupTestData(t, pool, ctx, accountID)

	// Create initial state
	initialState := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID,
		Key:       "test-settings",
		State:     []byte("initial-data"),
		Nonce:     []byte("nonce-1"),
		Version:   0,
	}
	err := repo.Upsert(ctx, initialState)
	require.NoError(t, err)
	require.Equal(t, int64(1), initialState.Version)

	// ACT: Update with correct version
	updatedState := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID,
		Key:       "test-settings",
		State:     []byte("updated-data"),
		Nonce:     []byte("nonce-2"),
		Version:   1, // Must match current version!
	}

	err = repo.Upsert(ctx, updatedState)

	// ASSERT: Should succeed and increment version
	require.NoError(t, err)
	assert.Equal(t, initialState.ID, updatedState.ID, "Should update same record")
	assert.Equal(t, int64(2), updatedState.Version, "Version should increment to 2")
}

// TestStateRepository_Upsert_VersionConflict tests optimistic locking failure
// This is the CRITICAL test - ensures conflicts are detected!
func TestStateRepository_Upsert_VersionConflict(t *testing.T) {
	pool := getTestPool(t)
	repo := NewPostgresEncryptedStateRepository(pool)
	accountRepo := NewPostgresAccountRepository(pool)
	deviceRepo := NewPostgresDeviceRepository(pool)
	ctx := context.Background()

	// Create test account and devices
	accountID, deviceID1 := setupTestAccountAndDevice(t, ctx, pool, accountRepo, deviceRepo)
	defer cleanupTestData(t, pool, ctx, accountID)

	// Create second device for the same account
	device2 := &models.Device{
		AccountID:  accountID,
		Name:       "Test Device 2",
		DeviceType: "desktop",
	}
	err := deviceRepo.Create(ctx, device2)
	require.NoError(t, err)
	deviceID2 := device2.ID

	// Create initial state (version 1)
	initialState := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID1,
		Key:       "test-settings",
		State:     []byte("initial-data"),
		Nonce:     []byte("nonce-1"),
		Version:   0,
	}
	err = repo.Upsert(ctx, initialState)
	require.NoError(t, err)

	// Simulate Device 1 updating (version 1 → 2)
	device1Update := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID1,
		Key:       "test-settings",
		State:     []byte("device1-update"),
		Nonce:     []byte("nonce-device1"),
		Version:   1, // Correct version
	}
	err = repo.Upsert(ctx, device1Update)
	require.NoError(t, err)
	require.Equal(t, int64(2), device1Update.Version)

	// ACT: Device 2 tries to update with stale version (thinks it's still version 1)
	device2Update := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID2,
		Key:       "test-settings",
		State:     []byte("device2-update"),
		Nonce:     []byte("nonce-device2"),
		Version:   1, // STALE! Current version is 2
	}

	err = repo.Upsert(ctx, device2Update)

	// ASSERT: Should return version conflict error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrVersionConflict, "Should detect version conflict")
}

// TestStateRepository_GetByKey tests retrieving state by account + key
func TestStateRepository_GetByKey(t *testing.T) {
	pool := getTestPool(t)
	repo := NewPostgresEncryptedStateRepository(pool)
	accountRepo := NewPostgresAccountRepository(pool)
	deviceRepo := NewPostgresDeviceRepository(pool)
	ctx := context.Background()

	// Create test account and device
	accountID, deviceID := setupTestAccountAndDevice(t, ctx, pool, accountRepo, deviceRepo)
	defer cleanupTestData(t, pool, ctx, accountID)

	// Create a state
	state := &models.EncryptedState{
		AccountID: accountID,
		DeviceID:  deviceID,
		Key:       "my-settings",
		State:     []byte("encrypted-data"),
		Nonce:     []byte("nonce"),
		Version:   0,
	}
	err := repo.Upsert(ctx, state)
	require.NoError(t, err)

	// ACT: Retrieve by key
	retrieved, err := repo.GetByKey(ctx, accountID, "my-settings")

	// ASSERT: Should find the state
	require.NoError(t, err)
	assert.Equal(t, state.ID, retrieved.ID)
	assert.Equal(t, "my-settings", retrieved.Key)
	assert.Equal(t, []byte("encrypted-data"), retrieved.State)
}

// Helper functions for test setup

// getTestPool returns a connection pool for testing
// In production, you'd use a test database URL from environment
func getTestPool(t *testing.T) *pgxpool.Pool {
	// TODO: Replace with your test database URL
	// For now, assumes same DB as dev (not ideal, but works)
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/edgesync?sslmode=disable")
	require.NoError(t, err, "Failed to connect to test database")
	return pool
}

// setupTestAccountAndDevice creates a test account and device for foreign key constraints
func setupTestAccountAndDevice(t *testing.T, ctx context.Context, pool *pgxpool.Pool, accountRepo *PostgresAccountRepository, deviceRepo *PostgresDeviceRepository) (uuid.UUID, uuid.UUID) {
	// Create test account
	account := &models.Account{
		Email:        "test-" + uuid.New().String() + "@example.com",
		PasswordHash: "test-hash",
	}
	err := accountRepo.Create(ctx, account)
	require.NoError(t, err, "Failed to create test account")

	// Create test device
	device := &models.Device{
		AccountID:  account.ID,
		Name:       "Test Device",
		DeviceType: "desktop",
	}
	err = deviceRepo.Create(ctx, device)
	require.NoError(t, err, "Failed to create test device")

	return account.ID, device.ID
}

// cleanupTestData removes test data (cascades to states and devices)
func cleanupTestData(t *testing.T, pool *pgxpool.Pool, ctx context.Context, accountID uuid.UUID) {
	// Delete account (cascades to devices and states due to ON DELETE CASCADE)
	accountRepo := NewPostgresAccountRepository(pool)
	err := accountRepo.Delete(ctx, accountID)
	if err != nil && err != ErrNotFound {
		t.Logf("Warning: failed to cleanup test account: %v", err)
	}
}
