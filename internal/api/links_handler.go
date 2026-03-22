package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	config "github.com/moklidia/go-project-278/internal/config"
	db "github.com/moklidia/go-project-278/internal/db"
	service "github.com/moklidia/go-project-278/internal/service"
)

type LinkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

type linkInput struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

func GetLinks(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var links []db.Link
		var err error
		var requestedRange *rangeBounds
		paginationRange := c.Query("range")

		if paginationRange != "" {
			pagination, bounds, err := getLimitAndOffsetFromQuery(paginationRange)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			requestedRange = &bounds

			links, err = queries.ListLinks(c.Request.Context(), db.ListLinksParams{
				Limit:  pagination.Limit,
				Offset: pagination.Offset,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			links, err = queries.ListAllLinks(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		total, err := queries.CountLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		contentRange := buildContentRange("links", requestedRange, len(links), total)
		c.Header("Content-Range", contentRange)

		responseLinks := make([]LinkResponse, 0, len(links))
		for _, link := range links {
			responseLinks = append(responseLinks, toLinkResponse(link))
		}

		c.JSON(http.StatusOK, responseLinks)
	}
}

func GetLink(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wrong id format"})
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		link, err := queries.GetLink(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, toLinkResponse(link))
	}
}

func CreateLink(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input linkInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		originalURL := input.OriginalURL
		if originalURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "original_url is required"})
			return
		}

		shortName := input.ShortName
		if shortName == "" {
			randomString, err := service.GenerateShortName(5)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short name"})
				return
			}
			shortName = randomString
		}

		appURL := config.GetEnv("APP_URL", "https://short.io")
		shortURL := fmt.Sprintf("%s/%s", appURL, shortName)
		created, err := queries.CreateLink(c.Request.Context(), db.CreateLinkParams{
			OriginalUrl: originalURL,
			ShortName:   shortName,
			ShortUrl:    shortURL,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				c.JSON(http.StatusConflict, gin.H{
					"error": "short name already exists",
					"code":  "short_name_conflict",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
			return
		}

		c.JSON(http.StatusCreated, toLinkResponse(created))
	}
}

func UpdateLink(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wrong id format"})
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var input linkInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		current, err := queries.GetLink(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		originalURL := input.OriginalURL
		shortName := input.ShortName

		if shortName == "" {
			shortName = current.ShortName
		}

		appURL := config.GetEnv("APP_URL", "https://short.io")
		shortURL := fmt.Sprintf("%s/%s", appURL, shortName)

		if originalURL == "" {
			originalURL = current.OriginalUrl
		}

		rowsAffected, err := queries.UpdateLink(c.Request.Context(), db.UpdateLinkParams{
			ID:          id,
			OriginalUrl: originalURL,
			ShortName:   shortName,
			ShortUrl:    shortURL,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				c.JSON(http.StatusConflict, gin.H{
					"error": "short name already exists",
					"code":  "short_name_conflict",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func DeleteLink(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wrong id format"})
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		rowsAffected, err := queries.DeleteLink(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})
	}
}

func toLinkResponse(l db.Link) LinkResponse {
	return LinkResponse{
		ID:          l.ID,
		OriginalURL: l.OriginalUrl,
		ShortName:   l.ShortName,
		ShortURL:    l.ShortUrl,
	}
}
