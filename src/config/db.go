package config

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectDB() *mongo.Database {
	ctx := context.Background()
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	MONGO_URI := os.Getenv("MONGO_URI")

	clientOptions := options.Client().ApplyURI(MONGO_URI)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	db := client.Database("webhook-delivery-system")

	pingerr := db.Client().Ping(ctx, nil)

	if err != nil {
		fmt.Printf("pingerr: %v\n", pingerr)
		os.Exit(1)
	}

	println("Connected to DB")

	return db
}
