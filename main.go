package main

import (
	"net/http"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config.ConnectDB()

	r.GET("/", healthRoute)

	r.Run(":3000")
}

func healthRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})
}
