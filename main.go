package main

import (
	"net/http"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/handlers"
	"github.com/arjav1528/webhook-delivery-system/src/queue"
	"github.com/arjav1528/webhook-delivery-system/src/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config.ConnectDB()
	queue.Init()
	worker.StartWorkers(5)

	r.GET("/", healthRoute)
	r.POST("/webhooks", handlers.RegisterWebhook)
	r.POST("/events", handlers.TriggerEvent)
	r.GET("/deliveries/:id", handlers.GetDeliveryStatus)

	r.Run(":3000")
}

func healthRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})
}
