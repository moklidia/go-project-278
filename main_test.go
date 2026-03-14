package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

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

func TestGetLinks(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures("testdata/links.yml")
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := setupRouter(queries)

	req,err := http.NewRequest("GET", "/api/links", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Links []LinkResponse `json:"links"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)

	require.NoError(t, err)

	links := response.Links
	require.Len(t, links, 3)

	assert.Equal(t, fixtureLinks[0].OriginalURL, links[0].OriginalURL)
	assert.Equal(t, fixtureLinks[1].OriginalURL, links[1].OriginalURL)
	assert.Equal(t, fixtureLinks[2].OriginalURL, links[2].OriginalURL)
}

func setupRouter(queries *db.Queries) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/links", getLinksHandler(queries))
	return router
}

func setupTx(t *testing.T) (pgx.Tx, *db.Queries) {
	conn := initTestDB(t)

	tx, err := conn.Begin(context.Background())
	require.NoError(t, err)

	q := db.New(tx)

	t.Cleanup(func() {
		err := tx.Rollback(context.Background())
		require.NoError(t, err)
	})

	return tx, q
}

func LoadLinkFixtures(path string) ([]LinkFixture, error) {
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
