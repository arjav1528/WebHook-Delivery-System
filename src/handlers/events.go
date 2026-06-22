package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/models"
	"github.com/arjav1528/webhook-delivery-system/src/queue"
	"github.com/gin-gonic/gin"
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

	log.Default().Println("event: ", event)

	res, err := config.EventCollection.InsertOne(c.Request.Context(), event)
	if err != nil {
		log.Default().Println("event insert err: ", string(err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Default().Println("res: ", res)

	var webhook []models.Webhook

	filter := bson.M{"event": event.Type}

	cur, err := config.WebHookCollection.Find(c.Request.Context(), filter)
	if err != nil {
		log.Default().Println("webhook find err: ", string(err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(c.Request.Context())

	if err := cur.All(c.Request.Context(), &webhook); err != nil {
		log.Default().Println("webhook find all err: ", string(err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, hook := range webhook {
		newDelivery := models.Delivery{
			WebhookID: hook.ID,
			EventID:   event.ID,
			Status:    models.DeliveryPending,
			Retry:     0,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := newDelivery.Validate(); err != nil {
			log.Default().Println("delivery validate err: ", string(err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		deliveryRes, err := config.DeliveryCollection.InsertOne(c.Request.Context(), newDelivery)
		if err != nil {
			log.Default().Println("delivery insert err: ", string(err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		deliveryID, ok := deliveryRes.InsertedID.(bson.ObjectID)
		if !ok {
			log.Default().Println("delivery insert returned an unexpected id type")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "delivery insert returned an unexpected id type",
			})
			return
		}

		deliveryJob := queue.DeliveryJob{
			DeliveryID:    deliveryID,
			WebhookID:     hook.ID,
			EventID:       event.ID,
			RetryCount:    0,
			NextRetryTime: time.Now().UTC(),
		}

		if err := queue.EnqueueJob(deliveryJob); err != nil {
			log.Default().Println("queue enqueue job err: ", string(err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to enqueue job: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "event accepted",
		"event_id": event.ID.Hex(),
		"webhooks": len(webhook),
	})
}
