package main

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func initTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/link_shortener_test?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	err = goose.Up(sqlDB, "db/migrations")
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}
