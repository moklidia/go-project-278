package api

import (
	"github.com/gin-contrib/logger"
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	db "github.com/moklidia/go-project-278/internal/db"
)

func SetupRouter(queries *db.Queries) *gin.Engine {
	router := gin.New()

	if gin.Mode() != gin.TestMode {
		router.Use(logger.SetLogger())
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
			WaitForDelivery: true,
		}))
		router.Use(cors.New(cors.Config{
			AllowOrigins: []string{"http://localhost:5173"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
		}))	
	}

	router.TrustedPlatform = gin.PlatformCloudflare

	router.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

	router.GET("/api/links", GetLinks(queries))
	router.GET("/api/links/:id", GetLink(queries))
	router.POST("/api/links", CreateLink(queries))
	router.PUT("/api/links/:id", UpdateLink(queries))
	router.DELETE("/api/links/:id", DeleteLink(queries))
	router.GET("/r/:code", RedirectToLink(queries))

	return router
}
