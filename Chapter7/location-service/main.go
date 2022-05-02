package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const locationIdFormat = "location:%s"

type Location struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
}

type LocationNearMe struct {
	Location Location `json:"location"`
	Distance float64  `json:"distance"`
}

var (
	client = redis.NewClient(&redis.Options{
		Addr:     getStrEnv("REDIS_HOST", "localhost:6379"),
		Password: getStrEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	})
)

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/location/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")

		if location, err := fetchLocation(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"id": id, "message": err.Error()})
		} else if location == nil {
			c.JSON(http.StatusNotFound, gin.H{"id": id, "message": "not found"})
		} else {
			c.JSON(http.StatusOK, gin.H{"location": location})
		}

	})

	r.POST("/location", func(c *gin.Context) {
		var location Location

		if err := c.BindJSON(&location); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"location": location, "created": false, "message": err.Error()})
			return
		}

		if err := persistLocation(c, location); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"location": location, "created": false, "message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"location": location, "created": true, "message": "Location Created Successfully"})

	})

	r.GET("/location/nearby", func(c *gin.Context) {
		latitude, err := strconv.ParseFloat(c.Request.URL.Query().Get("latitude"), 64)

		if err != nil {
			badRequestResponse(c, "latitude", err)
			return
		}

		longitude, err := strconv.ParseFloat(c.Request.URL.Query().Get("longitude"), 64)

		if err != nil {
			badRequestResponse(c, "longitude", err)
			return
		}

		var unit string
		if value, exists := c.Params.Get("unit"); exists {
			unit = value
		} else {
			unit = "km"
		}

		distance, err := strconv.ParseFloat(c.Request.URL.Query().Get("distance"), 64)

		if err != nil {
			badRequestResponse(c, "distance", err)
			return
		}

		if locationsNearMe, err := nearByLocations(c, longitude, latitude, unit, distance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"locations": locationsNearMe})
		}

	})

	return r
}

func badRequestResponse(c *gin.Context, field string, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"field": field, "message": err.Error()})
}

func nearByLocations(c context.Context, longitude float64, latitude float64, unit string, distance float64) ([]LocationNearMe, error) {
	var locationsNearMe []LocationNearMe = make([]LocationNearMe, 0)
	query := &redis.GeoRadiusQuery{Unit: unit, WithDist: true, Radius: distance, Sort: "ASC"}
	geoRadius := client.GeoRadius(c, "locations", longitude, latitude, query)

	if err := geoRadius.Err(); err != nil {
		return nil, err
	}

	geoLocations, err := geoRadius.Result()

	if err != nil {
		return nil, err
	}

	for _, geoLocation := range geoLocations {
		if location, err := fetchLocation(c, geoLocation.Name); err != nil {
			return nil, err
		} else {
			locationsNearMe = append(locationsNearMe, LocationNearMe{
				Location: *location,
				Distance: geoLocation.Dist,
			})
		}
	}

	return locationsNearMe, nil
}

func fetchLocation(c context.Context, id string) (*Location, error) {
	hgetAll := client.HGetAll(c, fmt.Sprintf(locationIdFormat, id))

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

	latitude, _ := strconv.ParseFloat(ires["Latitude"], 64)
	longitude, _ := strconv.ParseFloat(ires["Longitude"], 64)

	location := Location{Id: ires["Id"], Name: ires["Name"], Description: ires["Description"], Longitude: longitude, Latitude: latitude}
	return &location, nil
}

func persistLocation(c context.Context, location Location) error {
	hmset := client.HSet(c,
		fmt.Sprintf(locationIdFormat, location.Id), "Id",
		location.Id, "Name",
		location.Name, "Description",
		location.Description, "Longitude",
		location.Longitude, "Latitude",
		location.Latitude)

	if hmset.Err() != nil {
		return hmset.Err()
	}

	geoLoc := &redis.GeoLocation{Longitude: location.Longitude, Latitude: location.Latitude, Name: location.Id}

	gadd := client.GeoAdd(c, "locations", geoLoc)

	if gadd.Err() != nil {
		return gadd.Err()
	}

	return nil
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
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
