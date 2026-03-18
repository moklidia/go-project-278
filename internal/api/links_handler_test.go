package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/gin-gonic/gin"

	db "github.com/moklidia/go-project-278/internal/db"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestGetLinks(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

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

func TestGetLink(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	links, err := queries.ListLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	req,err := http.NewRequest("GET", fmt.Sprintf("/api/links/%d", link.ID), nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Link LinkResponse `json:"link"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)

	require.NoError(t, err)

	assert.Equal(t, link.ID, response.Link.ID)
	assert.Equal(t, link.OriginalUrl, response.Link.OriginalURL)
}

func TestCreateLink(t *testing.T) {
	_, queries := setupTx(t)
	router := SetupRouter(queries)
	body, err := json.Marshal(map[string]string{
		"original_url": "https://telegram.com",
		"short_name":   "telegram",
	})
	require.NoError(t, err)

	req,err := http.NewRequest("POST", "/api/links", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Link LinkResponse `json:"link"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Greater(t, response.Link.ID, int64(0))
	assert.Equal(t, "https://telegram.com", response.Link.OriginalURL)
	assert.Equal(t, "telegram", response.Link.ShortName)
}

func TestUpdateLink(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	links, err := queries.ListLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	body, err := json.Marshal(map[string]string{
		"original_url": "https://telegram.com",
		"short_name":   "telegram",
	})
	require.NoError(t, err)

	req,err := http.NewRequest("PUT", fmt.Sprintf("/api/links/%d", link.ID), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Link LinkResponse `json:"link"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	updated, err := queries.GetLink(context.Background(), link.ID)
	require.NoError(t, err)

	assert.Equal(t, "https://telegram.com", updated.OriginalUrl)
	assert.Equal(t, "telegram", updated.ShortName)
	assert.Equal(t, "https://short.io/telegram", updated.ShortUrl)
}

func TestDeleteLink(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	links, err := queries.ListLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	req,err := http.NewRequest("DELETE", fmt.Sprintf("/api/links/%d", link.ID), nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	_, err = queries.GetLink(context.Background(), link.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestUpdateLinkNotFound(t *testing.T) {
	_, queries := setupTx(t)
	router := SetupRouter(queries)

	body, err := json.Marshal(map[string]string{
		"original_url": "https://telegram.com",
		"short_name":   "telegram",
	})
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/api/links/999999", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLinkNotFound(t *testing.T) {
	_, queries := setupTx(t)
	router := SetupRouter(queries)

	req, err := http.NewRequest("DELETE", "/api/links/999999", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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

