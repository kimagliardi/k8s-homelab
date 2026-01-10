package main

import (
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)
		v1.GET("/healthz", healthzHandler)
	}

	return r
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func healthzHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func predictHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func main() {
	router := setupRouter()
	router.Run(":8080")
}
