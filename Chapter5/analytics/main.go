package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var (
	client = redis.NewClient(&redis.Options{
		Addr:     getStrEnv("REDIS_HOST", "localhost:6379"),
		Password: getStrEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	})
)

func getIntEnv(key string, defaultvaule int) int {
	if value := os.Getenv(key); len(value) == 0 {
		return defaultvaule
	} else {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		} else {
			return defaultvaule
		}
	}
}

func getStrEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); len(value) == 0 {
		return defaultValue
	} else {
		return value
	}
}

func main() {

	stream := "task-stream"
	consumerGroup := "analytics-group"

	consumer := "analytics-consumer"

	ctx := context.Background()

	client.XGroupCreate(ctx, stream, consumerGroup, "0").Result()

	for {
		entries, err := client.XReadGroup(ctx,
			&redis.XReadGroupArgs{
				Group:    consumerGroup,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    1,
				Block:    0,
				NoAck:    false,
			},
		).Result()

		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < len(entries[0].Messages); i++ {
			messageID := entries[0].Messages[i].ID
			values := entries[0].Messages[i].Values

			taskId := values["task_id"]
			timestamp := values["timestamp"]
			locationId := values["location_id"]

			log.Printf("Received %v %v %v", taskId, timestamp, locationId)

			client.XAck(ctx, stream, consumerGroup, messageID)
		}
	}

}
