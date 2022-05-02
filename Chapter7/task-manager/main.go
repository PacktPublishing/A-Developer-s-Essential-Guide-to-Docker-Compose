package main

import (
	"net/http"
	"os"
	"strconv"
	"task-manager/location"
	"task-manager/stream"
	"task-manager/task"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	client = redis.NewClient(&redis.Options{
		Addr:     getStrEnv("REDIS_HOST", "localhost:6379"),
		Password: getStrEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	})

	locationService = location.LocationService{
		LocationServiceEndpoint: getStrEnv("LOCATION_HOST", "http://localhost:8081"),
	}

	taskStream = stream.TaskStream{
		Client: client,
	}

	taskService = task.TaskService{
		Client:          client,
		LocationService: &locationService,
		TaskStream:      &taskStream,
	}
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get tasks
	r.GET("/task", func(c *gin.Context) {
		if tasks, err := taskService.FetchTasks(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"tasks": tasks})
		}

	})

	// Get task
	r.GET("/task/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")

		if task, err := taskService.FetchTask(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"id": id, "message": err.Error()})
		} else if task == nil {
			c.JSON(http.StatusNotFound, gin.H{"id": id, "message": "not found"})
		} else {
			if task.Location != nil {
				if locationsNearMe, err := locationService.FindLocationNearMe(task.Location.Longitude, task.Location.Latitude, "km", 1.0); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"id": id, "message": err.Error()})
				} else {
					locationsNearMe := excludeSelf(task.Location.Id, locationsNearMe)
					c.JSON(http.StatusOK, gin.H{"task": task, "locationsNearMe": locationsNearMe})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{"task": task})
			}
		}
	})

	// Add task
	r.POST("/task", func(c *gin.Context) {
		var task task.Task

		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"task": task, "created": false, "message": err.Error()})
			return
		}

		if err := taskService.PersistTask(c, task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"task": task, "created": false, "message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"task": task, "created": true, "message": "Task Created Successfully"})
	})

	r.DELETE("/task/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		if err := taskService.DeleteTask(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"id": id, "message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"id": id, "message": "Task deleted"})
		}

	})

	return r
}

func excludeSelf(id string, locations []location.LocationNearMe) []location.LocationNearMe {
	if len(locations) > 0 && locations[0].Location.Id == id {
		return locations[1:]
	} else {
		return locations
	}
}

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
	r := setupRouter()
	r.Run(getStrEnv("TASK_MANAGER_HOST", ":8080"))
}
