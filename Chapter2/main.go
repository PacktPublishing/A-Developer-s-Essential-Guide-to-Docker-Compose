package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
}

var taskMap = make(map[string]Task)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get tasks
	r.GET("/task", func(c *gin.Context) {
		tasks := []Task{}
		for _, v := range taskMap {
			tasks = append(tasks, v)
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	})

	// Get task
	r.GET("/task/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		task, ok := taskMap[id]
		if ok {
			c.JSON(http.StatusOK, gin.H{"task": task})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"id": id, "message": "not found"})
		}
	})

	// Add task
	r.POST("/task", func(c *gin.Context) {
		var task Task

		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusOK, gin.H{"task": task, "created": false, "message": err.Error()})
		} else {
			taskMap[task.Id] = task
			c.JSON(http.StatusCreated, gin.H{"task": task, "created": true, "message": "Task Created Successfully"})
		}

	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
