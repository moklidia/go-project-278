package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLinks(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	req, err := http.NewRequest("GET", "/api/links", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []LinkResponse

	err = json.Unmarshal(w.Body.Bytes(), &response)

	require.NoError(t, err)
	require.Len(t, response, 3)

	assert.Equal(t, fixtureLinks[0].OriginalURL, response[0].OriginalURL)
	assert.Equal(t, fixtureLinks[1].OriginalURL, response[1].OriginalURL)
	assert.Equal(t, fixtureLinks[2].OriginalURL, response[2].OriginalURL)
}

func TestGetLinksWithPagination(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	req, err := http.NewRequest("GET", "/api/links?range=[1,2]", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 1-1/3", w.Header().Get("Content-Range"))

	var response []LinkResponse
	
	err = json.Unmarshal(w.Body.Bytes(), &response)

	require.NoError(t, err)

	require.Len(t, response, 1)

	assert.Equal(t, fixtureLinks[1].OriginalURL, response[0].OriginalURL)
}

func TestGetLink(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	links, err := queries.ListAllLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/links/%d", link.ID), nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response LinkResponse

	err = json.Unmarshal(w.Body.Bytes(), &response)

	require.NoError(t, err)

	assert.Equal(t, link.ID, response.ID)
	assert.Equal(t, link.OriginalUrl, response.OriginalURL)
}

func TestCreateLink(t *testing.T) {
	_, queries := setupTx(t)
	router := SetupRouter(queries)
	body, err := json.Marshal(map[string]string{
		"original_url": "https://telegram.com",
		"short_name":   "telegram",
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/links", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response LinkResponse

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Greater(t, response.ID, int64(0))
	assert.Equal(t, "https://telegram.com", response.OriginalURL)
	assert.Equal(t, "telegram", response.ShortName)
}

func TestUpdateLink(t *testing.T) {
	_, queries := setupTx(t)
	fixtureLinks, err := LoadLinkFixtures(t)
	require.NoError(t, err)
	err = SeedLinks(context.Background(), queries, fixtureLinks)
	require.NoError(t, err)
	router := SetupRouter(queries)

	links, err := queries.ListAllLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	body, err := json.Marshal(map[string]string{
		"original_url": "https://telegram.com",
		"short_name":   "telegram",
	})
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/links/%d", link.ID), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response LinkResponse

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

	links, err := queries.ListAllLinks(context.Background())
	require.NoError(t, err)
	link := links[0]

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/links/%d", link.ID), nil)
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
