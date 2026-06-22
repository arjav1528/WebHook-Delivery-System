package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeliveryJob struct {
	DeliveryID    primitive.ObjectID
	WebhookID     primitive.ObjectID
	EventID       primitive.ObjectID
	RetryCount    int
	NextRetryTime time.Time
}

var RedisClient *redis.Client

const QueueKey = "queue:deliveries"

func Init() {
	RedisClient = config.RDB
}

func EnqueueJob(job DeliveryJob) error {
	ctx := context.Background()

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	if err = RedisClient.RPush(ctx, QueueKey, jobJSON).Err(); err != nil {
		return err
	}

	return nil
}

func DequeueJob() (DeliveryJob, error) {
	ctx := context.Background()

	var job DeliveryJob

	result, err := RedisClient.BLPop(ctx, 0, QueueKey).Result()
	if err != nil {
		return DeliveryJob{}, err
	}

	if len(result) < 2 {
		return job, nil
	}

	jobJSON := result[1]

	err = json.Unmarshal([]byte(jobJSON), &job)
	if err != nil {
		return job, err
	}

	return job, nil
}

func GetQueueLength() int64 {
	ctx := context.Background()
	length, err := RedisClient.LLen(ctx, QueueKey).Result()
	if err != nil {
		return 0
	}
	return length
}
