package main

import (
  "net/http"
	"log"

  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/logger"
)

func main() {
  router := gin.Default()
	router.Use(logger.SetLogger())
	router.Use(gin.Recovery())

  router.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

  if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
