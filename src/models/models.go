package models

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	EventType      string
	DeliveryStatus string
)

const (
	EventUserCreated EventType = "user.created"
	EventUserDeleted EventType = "user.deleted"
)

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySuccess DeliveryStatus = "success"
	DeliveryFailed  DeliveryStatus = "failed"
)

type Webhook struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	URL       string             `bson:"url,omitempty"`
	Event     EventType          `bson:"event"`
	CreatedAt time.Time          `bson:"created_at"`
}

type Event struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Type      EventType          `bson:"type"`
	Data      any                `bson:"data"`
	CreatedAt time.Time          `bson:"created_at"`
}

type Delivery struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	EventID   primitive.ObjectID `bson:"event_id"`
	WebhookID primitive.ObjectID `bson:"webhook_id"`
	Status    DeliveryStatus     `bson:"status"`
	Retry     int                `bson:"retry"`
	LastError string             `bson:"last_error,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (et EventType) IsValid() bool {
	switch et {
	case EventUserCreated, EventUserDeleted:
		return true
	default:
		return false
	}
}

func (ds DeliveryStatus) IsValid() bool {
	switch ds {
	case DeliveryPending, DeliverySuccess, DeliveryFailed:
		return true
	default:
		return false
	}
}

func (d *Delivery) Validate() error {
	if d.EventID.IsZero() {
		return fmt.Errorf("event_id is required")
	}

	if d.WebhookID.IsZero() {
		return fmt.Errorf("webhook_id is required")
	}

	if !d.Status.IsValid() {
		return fmt.Errorf("invalid status: %s. allowed: pending, success, failed", d.Status)
	}

	if d.Retry < 0 || d.Retry > 3 {
		return fmt.Errorf("retry count must be between 0 and 3, got %d", d.Retry)
	}

	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}

	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = d.CreatedAt
	}

	return nil
}

func (e *Event) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("event type is required")
	}

	if !e.Type.IsValid() {
		return fmt.Errorf("invalid event type: %s. allowed: user.created, user.deleted", e.Type)
	}

	if e.Data == nil {
		return fmt.Errorf("event data is required")
	}

	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}

	return nil
}

func (w *Webhook) Validate() error {
	if w.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if !strings.HasPrefix(w.URL, "http://") && !strings.HasPrefix(w.URL, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if w.Event == "" {
		return fmt.Errorf("event type is required")
	}

	if !w.Event.IsValid() {
		return fmt.Errorf("invalid event type: %s. allowed: user.created, user.deleted", w.Event)
	}

	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now().UTC()
	}

	return nil
}
