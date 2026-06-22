package handlers

import (
	"errors"
	"net/http"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetDeliveryStatus(c *gin.Context) {
	deliveryID := c.Param("id")

	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "delivery ID is required"})
		return
	}

	var delivery models.Delivery

	id, err := bson.ObjectIDFromHex(deliveryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery ID"})
		return
	}

	if err := config.DeliveryCollection.FindOne(c.Request.Context(), bson.M{"_id": id}).Decode(&delivery); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "delivery not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"delivery_id": delivery.ID.Hex(),
		"status":      delivery.Status,
		"retry_count": delivery.Retry,
		"last_error":  delivery.LastError,
		"created_at":  delivery.CreatedAt,
		"updated_at":  delivery.UpdatedAt,
	})
}
