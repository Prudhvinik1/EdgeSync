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

const sessionPrefix = "session:"
const accountSessionsPrefix = "account:%s:sessions"

type RedisSessionRepository struct {
	client *redis.Client
}

func NewRedisSessionRepository(client *redis.Client) *RedisSessionRepository {
	return &RedisSessionRepository{client: client}

}

func (r *RedisSessionRepository) Create(ctx context.Context, session *models.Session) error {
	// 1. Serialize session to JSON
	jsonData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	// 2. Calculate TTL from session.ExpiresAt
	ttl := time.Until(session.ExpiresAt)

	// 3. Store with key "session:{id}" and TTL
	key := fmt.Sprintf("%s%s", sessionPrefix, session.ID)

	//4 Put the session in Redis
	err = r.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	//5 Put the session in the account sessions set
	accountKey := fmt.Sprintf(accountSessionsPrefix, session.AccountID)
	err = r.client.SAdd(ctx, accountKey, session.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to add session to account sessions: %w", err)
	}
	return nil
}

func (r *RedisSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	// 1. Get from Redis with key "session:{id}"
	key := fmt.Sprintf("%s%s", sessionPrefix, id)
	// 2. Handle redis.Nil (not found)

	jsonData, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 3. Deserialize JSON to session
	var session models.Session
	err = json.Unmarshal([]byte(jsonData), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil

}

func (r *RedisSessionRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Session, error) {

	accountKey := fmt.Sprintf(accountSessionsPrefix, accountID)
	sessionIDs, err := r.client.SMembers(ctx, accountKey).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get account sessions: %w", err)
	}

	var sessions []*models.Session
	var expiredIDs []interface{}

	for _, id := range sessionIDs {
		jsonData, err := r.client.Get(ctx, fmt.Sprintf("%s%s", sessionPrefix, id)).Result()
		if err == redis.Nil {
			expiredIDs = append(expiredIDs, id)
			continue
		}

		if err != nil {
			fmt.Printf("failed to get session %s: %v", id, err)
			continue
		}

		var session models.Session
		err = json.Unmarshal([]byte(jsonData), &session)
		if err != nil {
			fmt.Printf("failed to unmarshal session %s: %v", id, err)
			continue
		}

		sessions = append(sessions, &session)
	}

	// Clean up expired sessions
	if len(expiredIDs) > 0 {
		err = r.client.SRem(ctx, accountKey, expiredIDs).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to remove expired sessions: %w", err)
		}
	}
	return sessions, nil
}

func (r *RedisSessionRepository) Delete(ctx context.Context, id string) error {
	// Delete the session key
	// Hint: r.client.Del(ctx, key)
	key := fmt.Sprintf("%s%s", sessionPrefix, id)

	session, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	//5 Delete the session from the account sessions set
	accountKey := fmt.Sprintf(accountSessionsPrefix, session.AccountID)
	err = r.client.SRem(ctx, accountKey, id).Err()
	if err != nil {
		return fmt.Errorf("failed to remove session from account sessions: %w", err)
	}

	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) DeleteAllForAccount(ctx context.Context, accountID uuid.UUID) error {
	// Delete all sessions for the account

	accountKey := fmt.Sprintf(accountSessionsPrefix, accountID)
	sessionIDs, err := r.client.SMembers(ctx, accountKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get account sessions: %w", err)
	}
	for _, id := range sessionIDs {
		err = r.Delete(ctx, id)
		if err != nil {
			fmt.Printf("failed to delete session: %s\n", err)
			continue
		}
	}
	return nil
}
