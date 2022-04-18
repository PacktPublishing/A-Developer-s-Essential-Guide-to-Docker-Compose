package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
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

func pushProcessingDurationToPrometheus(duration time.Duration) {
	processingTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "task_event_process_duration",
		Help: "The timestamp of the last successful completion of a DB backup.",
	})

	millis := float64(duration.Milliseconds())
	processingTime.Set(millis)

	if err := push.New(getStrEnv("PUSH_GATEWAY", "http://localhost:9091"), "event_service").
		Collector(processingTime).
		Grouping("db", "customers").
		Push(); err != nil {
		fmt.Println("Could not push completion time to Pushgateway:", err)
	}
}

func main() {

	stream := "task-stream"
	consumerGroup := "analytics-group"

	consumer := "analytics-consumer"

	ctx := context.Background()

	client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0").Result()

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

			start := time.Now()
			log.Printf("Received %v %v %v", taskId, timestamp, locationId)

			client.XAck(ctx, stream, consumerGroup, messageID)
			elapsed := time.Since(start)

			pushProcessingDurationToPrometheus(elapsed)
		}
	}

}
