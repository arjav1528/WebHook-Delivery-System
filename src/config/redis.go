package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	doOnce sync.Once
	RDB    *redis.Client
)

func ConnectRedis() {
	doOnce.Do(func() {
		if err := godotenv.Load(".env"); err != nil {
			initErr = err
			fmt.Printf("err: %v\n", err)
			os.Exit(1)
		}

		REDIS_ADDR := os.Getenv("REDIS_ADDR")
		REDIS_USERNAME := os.Getenv("REDIS_USERNAME")
		REDIS_PASSWORD := os.Getenv("REDIS_PASSWORD")

		rdb := redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Username: REDIS_USERNAME,
			Password: REDIS_PASSWORD,
			DB:       0,
		})

		RDB = rdb
	})
}
