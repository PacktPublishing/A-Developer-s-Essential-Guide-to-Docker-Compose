package task

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"task-manager/location"
	"task-manager/stream"

	"github.com/go-redis/redis/v8"
)

type TaskService struct {
	Client          *redis.Client
	LocationService *location.LocationService
	TaskStream      *stream.TaskStream
}

type Task struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Timestamp   int64              `json:"timestamp"`
	Location    *location.Location `json:"location"`
}

func (ts *TaskService) PersistTask(c context.Context, task Task) error {

	values := []interface{}{"Id", task.Id, "Name", task.Name, "Description", task.Description, "Timestamp", task.Timestamp}

	if task.Location != nil {
		if err := ts.LocationService.AddLocation(task.Location); err != nil {
			return err
		}
		values = append(values, "location", task.Location.Id)
	}

	hmset := ts.Client.HSet(c, fmt.Sprintf("task:%s", task.Id), values)

	if hmset.Err() != nil {
		return hmset.Err()
	}

	z := redis.Z{Score: float64(task.Timestamp), Member: task.Id}
	zadd := ts.Client.ZAdd(c, "tasks", &z)

	if zadd.Err() != nil {
		return hmset.Err()
	}

	mes := stream.CreateTaskMessage(task.Id, task.Location, task.Timestamp)

	return ts.TaskStream.Publish(c, mes)
}

func (ts *TaskService) FetchTask(c context.Context, id string) (*Task, error) {
	hgetAll := ts.Client.HGetAll(c, fmt.Sprintf("task:%s", id))

	if err := hgetAll.Err(); err != nil {
		return nil, err
	}

	ires, err := hgetAll.Result()

	if err != nil {
		return nil, err
	}

	if l := len(ires); l == 0 {
		return nil, nil
	}

	timestamp, _ := strconv.ParseInt(ires["Timestamp"], 10, 64)

	task := Task{Id: ires["Id"], Name: ires["Name"], Description: ires["Description"], Timestamp: timestamp}

	if locationId, exists := ires["location"]; exists {
		if location, err := ts.LocationService.FindLocation(locationId); err != nil {
			return nil, err
		} else if location != nil {
			log.Default().Println("location found " + location.Id)
			task.Location = location
		} else {
			task.Location = nil
		}
	}

	return &task, nil
}

func (ts *TaskService) DeleteTask(c context.Context, id string) error {
	if err := ts.Client.Unlink(c, fmt.Sprintf("task:%s", id)).Err(); err != nil {
		return err
	}

	if err := ts.Client.ZRem(c, "tasks", id).Err(); err != nil {
		return err
	}

	return nil
}

func (ts *TaskService) FetchTasks(c context.Context) ([]*Task, error) {
	var tasks []*Task = make([]*Task, 0)

	zRange := ts.Client.ZRange(c, "tasks", 0, -1)

	if err := zRange.Err(); err != nil {
		return nil, err
	}

	ids, err := zRange.Result()

	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		if task, err := ts.FetchTask(c, id); err != nil {
			return nil, err
		} else {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}
