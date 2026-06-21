package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	once               sync.Once
	client             *mongo.Client
	db                 *mongo.Database
	WebHookCollection  *mongo.Collection
	EventCollection    *mongo.Collection
	DeliveryCollection *mongo.Collection
	initErr            error
)

func ConnectDB() *mongo.Database {
	once.Do(func() {
		ctx := context.Background()
		if err := godotenv.Load(".env"); err != nil {
			initErr = err
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}

		MONGO_URI := os.Getenv("MONGO_URI")

		clientOptions := options.Client().ApplyURI(MONGO_URI)

		mongoclient, err := mongo.Connect(clientOptions)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}
		client = mongoclient

		db = mongoclient.Database("webhook-delivery-system")

		pingerr := db.Client().Ping(ctx, nil)

		if pingerr != nil {
			fmt.Printf("pingerr: %v\n", pingerr)
			os.Exit(1)
		}

		println("Connected to DB")

		WebHookCollection = db.Collection("webhooks")
		EventCollection = db.Collection("events")
		DeliveryCollection = db.Collection("deliveries")
	})
	return db
}
