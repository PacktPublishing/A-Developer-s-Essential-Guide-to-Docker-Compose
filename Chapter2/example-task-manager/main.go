package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var taskMap = make(map[string]Task)

type Task struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/task", func(c *gin.Context) {
		tasks := []Task{}
		for _, v := range taskMap {
			tasks = append(tasks, v)
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	})

	r.GET("/task/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		task, ok := taskMap[id]
		if ok {
			c.JSON(http.StatusOK, gin.H{"task": task})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"id": id, "message": "not found"})
		}
	})

	r.POST("/task", func(c *gin.Context) {
		var task Task
		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusOK, gin.H{"task": task, "created": false, "message": err.Error()})
		} else {
			taskMap[task.Id] = task
			c.JSON(http.StatusCreated, gin.H{"task": task, "created": true, "message": "Task Created Successfully"})
		}
	})

	r.DELETE("/task/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		delete(taskMap, id)
		c.JSON(http.StatusOK, gin.H{"id": id, "message": "deleted"})
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
