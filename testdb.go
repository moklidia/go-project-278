package main

import (
	"os"
	"testing"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"database/sql"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func initTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_DATABASE_URL")

	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	err = goose.Up(sqlDB, "db/migrations")
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}
