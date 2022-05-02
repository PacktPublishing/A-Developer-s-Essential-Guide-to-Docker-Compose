package location

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	getLocationEndpoint     = "%s/location/%s"
	locationEndpoint        = "%s/location"
	nearbyLocationsEndpoint = "%s/location/nearby?longitude=%f&latitude=%f&distance=%f&unit=%s"
)

type LocationService struct {
	LocationServiceEndpoint string
}

type Location struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
}

type LocationResponse struct {
	Location Location `json:"location"`
}

type LocationNearMe struct {
	Location Location `json:"location"`
	Distance float64  `json:"distance"`
}

type LocationsNearMe struct {
	Locations []LocationNearMe `json:"locations"`
}

func (ls *LocationService) FindLocation(id string) (*Location, error) {
	endpoint := fmt.Sprintf(getLocationEndpoint, ls.LocationServiceEndpoint, id)

	var locationResponse LocationResponse

	respose, err := http.Get(endpoint)

	if err != nil {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(respose.Body)

	log.Println(string(responseData))

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseData, &locationResponse); err != nil {
		return nil, err
	}

	return &locationResponse.Location, nil
}

func (ls *LocationService) AddLocation(location *Location) error {
	if fetchedLocation, err := ls.FindLocation(location.Id); fetchedLocation != nil {
		return nil
	} else if err != nil {
		return err
	}

	locationJson, err := json.Marshal(location)

	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf(locationEndpoint, ls.LocationServiceEndpoint)

	respose, err := http.Post(endpoint, "application/json", bytes.NewBuffer(locationJson))

	if err != nil {
		return err
	}

	if _, err := ioutil.ReadAll(respose.Body); err != nil {
		return err
	}

	return nil
}

func (ls *LocationService) FindLocationNearMe(longitude float64, latitude float64, unit string, distance float64) ([]LocationNearMe, error) {

	endpoint := fmt.Sprintf(nearbyLocationsEndpoint, ls.LocationServiceEndpoint, longitude, latitude, distance, unit)

	var locationsNearMe LocationsNearMe

	respose, err := http.Get(endpoint)

	if err != nil {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(respose.Body)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseData, &locationsNearMe); err != nil {
		return nil, err
	}

	return locationsNearMe.Locations, nil
}
