package noaa

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	stationsDir    = "noaa_data"
	stationsFile   = "noaa_stations.xml"
	stationListUrl = "http://w1.weather.gov/xml/current_obs/index.xml"
)

// Contains an Array of Stations
type StationList struct {
	XMLName  xml.Name  `xml:"wx_station_index"`
	Stations []Station `xml:"station"`
}

// Representation of a NOAA weather station. Each has an ID, Latitude and Longitude
type Station struct {
	Id        string  `xml:"station_id"`
	Latitude  float64 `xml:"latitude"`
	Longitude float64 `xml:"longitude"`
}

// Creates or re-writes an existing NOAA Station List
// This an expensive call, as the file is over one megabyte.
// It is best to only call this manually when you need a new list.
func RefreshNOAAStationList() ([]byte, error) {
	resp, err := http.Get(stationListUrl)

	// Something is wrong with the remote API. Panic
	if err != nil {
		fmt.Println("Could not connect to API")
		fmt.Println(err)
	}

	defer resp.Body.Close()

	// Remove existing XML File
	os.Remove(filepath.Join(stationsDir, stationsFile))

	// Create Directories if necessary
	err = os.MkdirAll(filepath.Join(stationsDir), 0777)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Create Stations FIle
	file, err := os.Create(filepath.Join(stationsDir, stationsFile))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Write the xml results to the file
	body, _ := ioutil.ReadAll(resp.Body)
	file.Write(body)

	return body, nil
}

// Fetch a list for all NOAA stations
// Either uses the cached copy or downloads a new copy.
func Stations() *StationList {
	var body []byte

	// Check if there is a local cached of the stations
	file, err := ioutil.ReadFile(filepath.Join(stationsDir, stationsFile))

	if err != nil {
		// Fetch station list from API if we have no local file
		body, _ = RefreshNOAAStationList()
	} else {
		body = file
	}

	stationList := &StationList{}
	xml.Unmarshal(body, &stationList)

	return stationList
}

// Find the nearest station given a point
func (p *Point) NearestStation() *Station {
	closestStation := &Station{}
	closestDistance := 0.0
	stationList := Stations()

	for _, s := range stationList.Stations {
		p2 := &Point{Latitude: s.Latitude, Longitude: s.Longitude}
		distance := p.HaversineDistance(p2)

		if distance < closestDistance || closestDistance == 0.0 {
			closestDistance = distance
			*closestStation = s
		}
	}

	return closestStation
}
