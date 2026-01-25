package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prudhvinik1/edgesync/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	presenceKeyPrefix = "presence:"
	presenceTTL       = 60 * time.Second // Presence expires after 60 seconds without heartbeat
)

type RedisPresenceRepository struct {
	client *redis.Client
}

func NewRedisPresenceRepository(client *redis.Client) *RedisPresenceRepository {
	return &RedisPresenceRepository{client: client}
}

// SetPresence sets or updates the presence for a device with automatic TTL.
// Clients should call this every 30 seconds to maintain "online" status.
func (r *RedisPresenceRepository) SetPresence(ctx context.Context, presence *models.Presence) error {
	// Update LastSeen to now
	presence.LastSeen = time.Now()

	data, err := json.Marshal(presence)
	if err != nil {
		return fmt.Errorf("failed to marshal presence: %w", err)
	}

	key := presenceKey(presence.DeviceID)
	err = r.client.Set(ctx, key, data, presenceTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set presence: %w", err)
	}

	return nil
}

func (r *RedisPresenceRepository) GetPresence(ctx context.Context, deviceID uuid.UUID) (*models.Presence, error) {
	key := presenceKey(deviceID)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// No presence = device is offline
		return &models.Presence{
			DeviceID: deviceID,
			Status:   string(models.StatusOffline),
			LastSeen: time.Time{}, // Zero time indicates unknown
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}

	var presence models.Presence
	if err := json.Unmarshal([]byte(data), &presence); err != nil {
		return nil, fmt.Errorf("failed to unmarshal presence: %w", err)
	}

	return &presence, nil
}

func (r *RedisPresenceRepository) DeletePresence(ctx context.Context, deviceID uuid.UUID) error {
	key := presenceKey(deviceID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete presence: %w", err)
	}

	return nil
}

// GetBulkPresence retrieves presence for multiple devices in a single call.
// This is efficient for getting presence of all devices in an account.
func (r *RedisPresenceRepository) GetBulkPresence(ctx context.Context, deviceIDs []uuid.UUID) (map[uuid.UUID]models.Presence, error) {
	if len(deviceIDs) == 0 {
		return make(map[uuid.UUID]models.Presence), nil
	}

	// Build keys
	keys := make([]string, len(deviceIDs))
	for i, id := range deviceIDs {
		keys[i] = presenceKey(id)
	}

	// MGet retrieves multiple keys in one round trip
	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk presence: %w", err)
	}

	presenceMap := make(map[uuid.UUID]models.Presence)

	for i, result := range results {
		deviceID := deviceIDs[i]

		if result == nil {
			// Device is offline
			presenceMap[deviceID] = models.Presence{
				DeviceID: deviceID,
				Status:   string(models.StatusOffline),
				LastSeen: time.Time{},
			}
			continue
		}

		data, ok := result.(string)
		if !ok {
			continue
		}

		var presence models.Presence
		if err := json.Unmarshal([]byte(data), &presence); err != nil {
			// If we can't unmarshal, treat as offline
			presenceMap[deviceID] = models.Presence{
				DeviceID: deviceID,
				Status:   string(models.StatusOffline),
				LastSeen: time.Time{},
			}
			continue
		}

		presenceMap[deviceID] = presence
	}

	return presenceMap, nil
}

// Helper: build Redis key for presence
func presenceKey(deviceID uuid.UUID) string {
	return presenceKeyPrefix + deviceID.String()
}

