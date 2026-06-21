package queue

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeliveryJob struct {
	WebhookID     primitive.ObjectID
	EventID       primitive.ObjectID
	RetryCount    int
	NextRetryTime time.Time
}

var DeliveryQueue chan DeliveryJob

func Init() {
	DeliveryQueue = make(chan DeliveryJob, 1000)
}
