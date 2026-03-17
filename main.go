package main

import (
	"os"
	"log"
	"time"
	"fmt"
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/moklidia/go-project-278/internal/db"
	api "github.com/moklidia/go-project-278/internal/api"
)

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

  router := api.SetupRouter(queries)

  if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
