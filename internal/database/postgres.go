package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxConns = 10
	MinConns = 2
	MaxConnLifetime = 10 * time.Minute
	MaxConnIdleTime = 5 * time.Minute
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(databaseURL)

	if err != nil {
		fmt.Println("Error parsing postgres config: ", err)
		return nil, err
	}

	// Configure the pool
	config.MaxConns = MaxConns
	config.MinConns = MinConns
	config.MaxConnLifetime = MaxConnLifetime
	config.MaxConnIdleTime = MaxConnIdleTime

	// Create the pool with config
	pool, err := pgxpool.NewWithConfig(ctx, config)


	if err != nil {
		fmt.Println("Error creating postgres pool: ", err)
		return nil, fmt.Errorf("error creating postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error pinging postgres pool: %w", err)
	}

	fmt.Println("Postgres pool created successfully")	

	return pool, nil
}