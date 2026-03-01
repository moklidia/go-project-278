package main

import (
  "net/http"
	"log"
	"time"
	"os"
	"crypto/rand"
	"fmt"
	"math/big"
	"context"
	"strconv"
	"errors"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/moklidia/go-project-278/internal/db"

)

type Link struct {
	ID int
	OriginalURL string
	ShortName string
	ShortURL string
}

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(2 * time.Second)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := db.New(pool)

  router := gin.Default()
	router.Use(logger.SetLogger())
	router.Use(gin.Recovery())
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true, WaitForDelivery: true}))

  router.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

	router.GET("/api/links", getLinksHandler(queries))
	router.GET("/api/links/:id", getLinkHandler(queries))
	router.POST("/api/links", postLinkHandler(queries))
	router.PUT("/api/links/:id", updateLinkHandler(queries))
	router.DELETE("/api/links/:id", deleteLinkHandler(queries))

  if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortName(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}

	return string(b), nil
}

func getLinksHandler(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		links, err := queries.ListLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"links": links})
	}
}

func getLinkHandler(queries *db.Queries) gin.HandlerFunc {
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
		c.JSON(http.StatusOK, gin.H{"link": link})
	}
}

func updateLinkHandler(queries *db.Queries) gin.HandlerFunc {
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

		current, err := queries.GetLink(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		originalURL := c.Query("original_url")
		shortName := c.Query("short_name")
		shortURL := fmt.Sprintf("%s/%s", os.Getenv("APP_URL"), shortName)

		if shortName == "" {
			shortName = current.ShortName
			shortURL = current.ShortUrl
		}

		if originalURL == "" {
			originalURL = current.OriginalUrl
	  }

		_, err = queries.UpdateLink(c.Request.Context(), db.UpdateLinkParams{
			ID: id,
			OriginalUrl: originalURL,
			ShortName: shortName,
			ShortUrl: shortURL,
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

		c.JSON(http.StatusOK, gin.H{})
	}
}

func deleteLinkHandler(queries *db.Queries) gin.HandlerFunc {
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
		err = queries.DeleteLink(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, gin.H{})
	}
}

func postLinkHandler(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalURL := c.Query("original_url")
		if originalURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "original_url is required"})
			return
		}

		shortName := c.Query("short_name")
		if shortName == "" {
			randomString, err := GenerateShortName(5)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short name"})
				return
			}
			shortName = randomString
		}

		shortURL := fmt.Sprintf("%s/%s", os.Getenv("APP_URL"), shortName)
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

		c.JSON(http.StatusCreated, gin.H{"link": created})
	}
}
