package main

import (
	"bufio"
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func initTestDB(t *testing.T) *pgxpool.Pool {
	loadEnvFile(t, ".env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, dbURL)

	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	err = goose.Up(sqlDB, "db/migrations")
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}

func loadEnvFile(t *testing.T, path string) {
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, file.Close())
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		value = strings.Trim(value, "\"")
		if os.Getenv(key) == "" {
			t.Setenv(key, value)
		}
	}

	require.NoError(t, scanner.Err())
}
