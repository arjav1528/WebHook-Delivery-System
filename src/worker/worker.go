package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arjav1528/webhook-delivery-system/src/config"
	"github.com/arjav1528/webhook-delivery-system/src/models"
	"github.com/arjav1528/webhook-delivery-system/src/queue"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func StartWorkers(numWorkers int) {
	for range numWorkers {
		go ProcessJobs()
	}
}

func ProcessJobs() {
	for {
		job, err := queue.DequeueJob()
		if err != nil {
			fmt.Printf("Error dequeuing job: %v\n", err)
			continue
		}

		if job.NextRetryTime.After(time.Now()) {
			time.Sleep(time.Until(job.NextRetryTime))
		}

		ProcessDeliveryJobs(job)
	}
}

func ProcessDeliveryJobs(job queue.DeliveryJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var webhook models.Webhook
	if err := config.WebHookCollection.FindOne(ctx, bson.M{"_id": job.WebhookID}).Decode(&webhook); err != nil {
		fmt.Printf("err: %v\n", err)
		if updateErr := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliveryFailed, err.Error(), job.RetryCount); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}

	var event models.Event
	if err := config.EventCollection.FindOne(ctx, bson.M{"_id": job.EventID}).Decode(&event); err != nil {
		fmt.Printf("err: %v\n", err)
		if updateErr := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliveryFailed, err.Error(), job.RetryCount); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}

	payload := map[string]interface{}{
		"event_id": job.EventID,
		"type":     event.Type,
		"data":     event.Data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		if updateErr := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliveryFailed, err.Error(), job.RetryCount); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}

	h := hmac.New(sha256.New, []byte(webhook.Secret))
	h.Write(payloadBytes)
	signature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		if updateErr := HandleDeliveryFailure(ctx, job, err.Error()); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		if updateErr := HandleDeliveryFailure(ctx, job, err.Error()); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		fmt.Println("Webhook Success POST")
		if updateErr := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliverySuccess, "", job.RetryCount); updateErr != nil {
			fmt.Printf("err: %v\n", updateErr)
		}
		return
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Printf("err: %v\n", readErr)
	}

	errorMsg := fmt.Sprintf("HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	if updateErr := HandleDeliveryFailure(ctx, job, errorMsg); updateErr != nil {
		fmt.Printf("err: %v\n", updateErr)
	}
}

func HandleDeliveryFailure(ctx context.Context, job queue.DeliveryJob, errorMsg string) error {
	maxRetries := 3

	if job.RetryCount < maxRetries {
		backoffSeconds := 1 << uint(job.RetryCount)
		backoffDuration := time.Duration(backoffSeconds) * time.Second

		newJob := job
		newJob.RetryCount++
		newJob.NextRetryTime = time.Now().Add(backoffDuration)
		if err := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliveryPending, errorMsg, newJob.RetryCount); err != nil {
			return err
		}
		queue.EnqueueJob(newJob)

		fmt.Printf("↻ Retry scheduled: webhook %s, attempt %d, backoff %s\n", job.WebhookID, newJob.RetryCount, backoffDuration)
		return nil
	}

	if err := UpdateDeliveryStatus(ctx, job.DeliveryID, models.DeliveryFailed, errorMsg, job.RetryCount); err != nil {
		return err
	}

	fmt.Printf("✗ Delivery failed: webhook %s, event %s, error: %s\n", job.WebhookID, job.EventID, errorMsg)
	return nil
}

func UpdateDeliveryStatus(ctx context.Context, deliveryID bson.ObjectID, status models.DeliveryStatus, errMsg string, retryCount int) error {
	filter := bson.M{"_id": deliveryID}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"retry":      retryCount,
			"last_error": errMsg,
			"updated_at": time.Now().UTC(),
		},
	}

	_, err := config.DeliveryCollection.UpdateOne(ctx, filter, update)
	return err
}
