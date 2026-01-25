package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prudhvinik1/edgesync/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionRepository_Create tests creating a session with TTL
func TestSessionRepository_Create(t *testing.T) {
	client := getTestRedisClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	defer cleanupTestSessions(t, client, ctx)

	accountID := uuid.New()
	deviceID := uuid.New()

	// ACT: Create a session
	session := &models.Session{
		ID:        "session-123",
		AccountID: accountID,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, session)

	// ASSERT: Should succeed
	require.NoError(t, err)

	// Verify session exists in Redis
	retrieved, err := repo.GetByID(ctx, "session-123")
	require.NoError(t, err)
	assert.Equal(t, accountID, retrieved.AccountID)
	assert.Equal(t, deviceID, retrieved.DeviceID)

	// Verify secondary index was created
	sessions, err := repo.ListByAccountID(ctx, accountID)
	require.NoError(t, err)
	assert.Len(t, sessions, 1, "Account should have 1 session")
	assert.Equal(t, "session-123", sessions[0].ID)
}

// TestSessionRepository_Expiration tests that expired sessions are cleaned up
// This tests the lazy cleanup mechanism
func TestSessionRepository_Expiration(t *testing.T) {
	client := getTestRedisClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	defer cleanupTestSessions(t, client, ctx)

	accountID := uuid.New()
	deviceID := uuid.New()

	// Create a session with very short TTL (1 second)
	session1 := &models.Session{
		ID:        "expired-session",
		AccountID: accountID,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(1 * time.Second), // Expires in 1 second
		CreatedAt: time.Now(),
	}
	err := repo.Create(ctx, session1)
	require.NoError(t, err)

	// Create another session that won't expire
	session2 := &models.Session{
		ID:        "valid-session",
		AccountID: accountID,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	err = repo.Create(ctx, session2)
	require.NoError(t, err)

	// Wait for first session to expire
	time.Sleep(2 * time.Second)

	// ACT: List sessions - should trigger lazy cleanup
	sessions, err := repo.ListByAccountID(ctx, accountID)

	// ASSERT: Should only return valid session, expired one cleaned up
	require.NoError(t, err)
	assert.Len(t, sessions, 1, "Should only have 1 valid session")
	assert.Equal(t, "valid-session", sessions[0].ID, "Should return the valid session")

	// Verify expired session is gone from Redis
	_, err = repo.GetByID(ctx, "expired-session")
	assert.ErrorIs(t, err, ErrNotFound, "Expired session should not exist")
}

// TestSessionRepository_Delete tests removing a session and cleaning up index
func TestSessionRepository_Delete(t *testing.T) {
	client := getTestRedisClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	defer cleanupTestSessions(t, client, ctx)

	accountID := uuid.New()
	deviceID := uuid.New()

	// Create a session
	session := &models.Session{
		ID:        "session-to-delete",
		AccountID: accountID,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.GetByID(ctx, "session-to-delete")
	require.NoError(t, err)

	// ACT: Delete the session
	err = repo.Delete(ctx, "session-to-delete")

	// ASSERT: Should succeed
	require.NoError(t, err)

	// Verify session is gone
	_, err = repo.GetByID(ctx, "session-to-delete")
	assert.ErrorIs(t, err, ErrNotFound, "Session should be deleted")

	// Verify it's removed from secondary index
	sessions, err := repo.ListByAccountID(ctx, accountID)
	require.NoError(t, err)
	assert.Len(t, sessions, 0, "Account should have no sessions")
}

// TestSessionRepository_DeleteAllForAccount tests removing all sessions for an account
func TestSessionRepository_DeleteAllForAccount(t *testing.T) {
	client := getTestRedisClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	defer cleanupTestSessions(t, client, ctx)

	accountID := uuid.New()
	deviceID := uuid.New()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &models.Session{
			ID:        uuid.New().String(),
			AccountID: accountID,
			DeviceID:  deviceID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Verify they exist
	sessions, err := repo.ListByAccountID(ctx, accountID)
	require.NoError(t, err)
	assert.Len(t, sessions, 3, "Should have 3 sessions")

	// ACT: Delete all sessions for account
	err = repo.DeleteAllForAccount(ctx, accountID)

	// ASSERT: Should succeed
	require.NoError(t, err)

	// Verify all are gone
	sessions, err = repo.ListByAccountID(ctx, accountID)
	require.NoError(t, err)
	assert.Len(t, sessions, 0, "Account should have no sessions")
}

// Helper functions for test setup

// getTestRedisClient returns a Redis client for testing
func getTestRedisClient(t *testing.T) *redis.Client {
	// TODO: Replace with your test Redis URL
	// For now, assumes same Redis as dev
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use DB 1 for tests (different from production DB 0)
	})

	// Test connection
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to test Redis")

	return client
}

// cleanupTestSessions removes test data
func cleanupTestSessions(t *testing.T, client *redis.Client, ctx context.Context) {
	// Clean up test sessions
	keys, err := client.Keys(ctx, "session:*").Result()
	if err != nil {
		t.Logf("Warning: failed to get keys: %v", err)
		return
	}

	if len(keys) > 0 {
		err = client.Del(ctx, keys...).Err()
		if err != nil {
			t.Logf("Warning: failed to cleanup test sessions: %v", err)
		}
	}

	// Clean up secondary indexes
	indexKeys, err := client.Keys(ctx, "account:*:sessions").Result()
	if err == nil && len(indexKeys) > 0 {
		client.Del(ctx, indexKeys...)
	}
}

