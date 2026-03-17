package api

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	
  "gopkg.in/yaml.v3"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	db "github.com/moklidia/go-project-278/internal/db"

)

type LinkFixture struct {
	ID          int64  `yaml:"id"`
	OriginalURL string `yaml:"original_url"`
	ShortName   string `yaml:"short_name"`
	ShortURL    string `yaml:"short_url"`
}

type LinkFixtures struct {
	Links []LinkFixture `yaml:"links"`
}


func initTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/link_shortener_test?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	err = goose.Up(sqlDB, migrationsPath(t))
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}

func LoadLinkFixtures(t *testing.T) ([]LinkFixture, error) {
	fixturesDir := fixturesPath(t)
	path := filepath.Clean(filepath.Join(fixturesDir, "links.yml"))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fixtures LinkFixtures

	err = yaml.Unmarshal(data, &fixtures)
	if err != nil {
		return nil, err
	}

	return fixtures.Links, nil
}

func SeedLinks(ctx context.Context, q *db.Queries, links []LinkFixture) error {
	for _, l := range links {
		_, err := q.CreateLink(ctx, db.CreateLinkParams{
			OriginalUrl: l.OriginalURL,
			ShortName:   l.ShortName,
			ShortUrl:    l.ShortURL,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func migrationsPath(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", "..", "db", "migrations"))
}

func fixturesPath(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", "..", "testdata"))
}

