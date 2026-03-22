package api

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
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

type LinkVisitFixture struct {
	ID        int64  `yaml:"id"`
	ShortName string `yaml:"short_name"`
	CreatedAt string `yaml:"created_at"`
	IP        string `yaml:"ip"`
	UserAgent string `yaml:"user_agent"`
	Referer   string `yaml:"referer"`
	Status    int32  `yaml:"status"`
}

type LinkVisitFixtures struct {
	LinkVisits []LinkVisitFixture `yaml:"link_visits"`
}

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/link_shortener_test?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(sqlDB, migrationsPath()); err != nil {
		log.Fatal(err)
	}

	testPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func setupTx(t *testing.T) (pgx.Tx, *db.Queries) {
	tx, err := testPool.Begin(context.Background())
	require.NoError(t, err)

	q := db.New(tx)

	t.Cleanup(func() {
		err := tx.Rollback(context.Background())
		require.NoError(t, err)
	})

	return tx, q
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

func LoadLinkVisitFixtures(t *testing.T) ([]LinkVisitFixture, error) {
	fixturesDir := fixturesPath(t)
	path := filepath.Clean(filepath.Join(fixturesDir, "link_visits.yml"))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fixtures LinkVisitFixtures

	err = yaml.Unmarshal(data, &fixtures)
	if err != nil {
		return nil, err
	}

	return fixtures.LinkVisits, nil
}

func SeedLinkVisits(ctx context.Context, q *db.Queries, linkVisits []LinkVisitFixture) error {
	for _, visit := range linkVisits {
		link, err := q.LinkByShortName(ctx, visit.ShortName)
		if err != nil {
			return err
		}

		_, err = q.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
			LinkID:    link.ID,
			Ip:        visit.IP,
			UserAgent: visit.UserAgent,
			Referer:   visit.Referer,
			Status:    visit.Status,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func fixturesPath(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", "..", "testdata"))
}

func migrationsPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("failed to resolve migrations path")
	}

	dir := filepath.Dir(filename)
	return filepath.Clean(filepath.Join(dir, "..", "..", "db", "migrations"))
}
