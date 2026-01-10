package main

import (
	"github.com/gin-gonic/gin"
) // Importa o pacote Gin)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return r
}

func main() {
	router := setupRouter()
	router.Run(":8080")
}
