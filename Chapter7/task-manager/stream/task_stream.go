package stream

import (
	"context"
	"task-manager/location"

	"github.com/go-redis/redis/v8"
)

type TaskStream struct {
	Client *redis.Client
}

type TaskMessage struct {
	taskId      string
	location_id string
	timestamp   int64
}

func CreateTaskMessage(taskId string, location *location.Location, timestamp int64) TaskMessage {
	taskMessage := TaskMessage{
		taskId:    taskId,
		timestamp: timestamp,
	}

	if location != nil {
		taskMessage.location_id = location.Id
	}

	return taskMessage
}

func (ts *TaskMessage) toXValues() map[string]interface{} {
	return map[string]interface{}{"task_id": ts.taskId, "timestamp": ts.timestamp, "location_id": ts.location_id}
}

func (ts *TaskStream) Publish(c context.Context, message TaskMessage) error {

	cmd := ts.Client.XAdd(c, &redis.XAddArgs{
		Stream: "task-stream",
		ID:     "*",
		Values: message.toXValues(),
	})

	if _, err := cmd.Result(); err != nil {
		return err
	}

	return nil
}
