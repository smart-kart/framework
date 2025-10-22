package pgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smart-kart/framework/env"
)

//nolint:gochecknoglobals // expected to be at global level
var _ds *DataSource

// DataSource represents the database connection pool
type DataSource struct {
	pool *pgxpool.Pool
}

// Init initializes the database connection pool
func Init(ctx context.Context) error {
	dbHost := env.Get(env.DBHost)
	dbPort := env.Get(env.DBPort)
	dbUser := env.Get(env.DBUser)
	dbPassword := env.Get(env.DBPassword)
	dbName := env.Get(env.DBName)

	fmt.Printf("DEBUG: DB config - host=%s port=%s user=%s dbname=%s\n", dbHost, dbPort, dbUser, dbName)

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	_ds = &DataSource{pool: pool}
	return nil
}

// GetDS returns the global datasource instance
func GetDS() *DataSource {
	return _ds
}

// GetPool returns the connection pool
func (ds *DataSource) GetPool() *pgxpool.Pool {
	return ds.pool
}

// Close closes the connection pool
func (ds *DataSource) Close() {
	if ds.pool != nil {
		ds.pool.Close()
	}
}
