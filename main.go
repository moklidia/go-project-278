package main

import (
  "net/http"
	"log"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/logger"
)

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://07ffea63ea0cf272dfcf249302f33c36@o4510963247546368.ingest.de.sentry.io/4510963256590416",
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
  router := gin.Default()
	router.Use(logger.SetLogger())
	router.Use(gin.Recovery())
	router.Use(sentrygin.New(sentrygin.Options{}))

  router.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

  if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
