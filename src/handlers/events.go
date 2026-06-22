package handlers

import (
	"net/http"
	"time"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/models"
	"github.com/arjav1528/webhook-delivery-system/src/queue"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TriggerEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	if err := event.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res, err := config.EventCollection.InsertOne(c.Request.Context(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	eventID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "event insert returned an unexpected id type",
		})
		return
	}

	var webhook []models.Webhook

	filter := bson.M{"event": event.Type}

	cur, err := config.WebHookCollection.Find(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(c.Request.Context())

	if err := cur.All(c.Request.Context(), &webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, hook := range webhook {
		newDelivery := models.Delivery{
			WebhookID: hook.ID,
			EventID:   eventID,
			Status:    models.DeliveryPending,
			Retry:     0,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := newDelivery.Validate(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		deliveryRes, err := config.DeliveryCollection.InsertOne(c.Request.Context(), newDelivery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		deliveryID, ok := deliveryRes.InsertedID.(primitive.ObjectID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "delivery insert returned an unexpected id type",
			})
			return
		}

		deliveryJob := queue.DeliveryJob{
			DeliveryID:    deliveryID,
			WebhookID:     hook.ID,
			EventID:       eventID,
			RetryCount:    0,
			NextRetryTime: time.Now().UTC(),
		}

		queue.EnqueueJob(deliveryJob)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "event accepted",
		"event_id": eventID.Hex(),
		"webhooks": len(webhook),
	})
}
