package handlers

import (
	"net/http"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterWebhook(c *gin.Context) {
	var webhook models.Webhook
	webhookCollection := config.WebHookCollection

	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Incomplete Input",
			"error":   err.Error(),
		})
		return
	}

	webhook.ID = primitive.NewObjectID()

	if err := webhook.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Input",
			"error":   err.Error(),
		})
		return
	}

	res, err := webhookCollection.InsertOne(c.Request.Context(), webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to register webhook",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook registered successfully",
		"id":      res.InsertedID,
	})
}
