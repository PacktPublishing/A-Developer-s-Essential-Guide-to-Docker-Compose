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
	processingTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "task_event_process_duration",
		Help: "Time it took to complete a task",
	})
	processedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_event_processing_total",
			Help: "How many tasks have been processed",
		},
		[]string{"task"},
	).WithLabelValues("task")
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

func pushProcessingDurationToPrometheus(processingTime prometheus.Gauge) {
	if err := push.New(getStrEnv("PUSH_GATEWAY", "http://localhost:9091"), "task_event_process_duration").
		Collector(processingTime).
		Grouping("db", "event-service").
		Push(); err != nil {
		fmt.Println("Could not push completion time to Pushgateway:", err)
	}
}

func pushProcessingCount(processedCounter prometheus.Counter) {
	if err := push.New(getStrEnv("PUSH_GATEWAY", "http://localhost:9091"), "task_event_processing_total").
		Collector(processedCounter).
		Grouping("db", "event-service").
		Push(); err != nil {
		fmt.Println("Could not push tasks processed to Pushgateway:", err)
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

			processedCounter.Add(1)

			millis := float64(elapsed.Milliseconds())
			processingTime.Set(millis)

			pushProcessingDurationToPrometheus(processingTime)
			pushProcessingCount(processedCounter)
		}
	}

}
