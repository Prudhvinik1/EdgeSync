package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Ping the client to ensure connection is established
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("error pinging redis: %w", err)
	}

	fmt.Println("Redis client created successfully")

	return client, nil
}