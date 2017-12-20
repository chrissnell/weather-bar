package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const wgAPIBaseURL = "https://api.wunderground.com/api/"

// GeoLookup encapsulates WU's geolocation API response object
type GeoLookup struct {
	Location Location `json:"location"`
}

// Location contains info related to the geolocation request
type Location struct {
	City           string         `json:"city"`
	State          string         `json:"state"`
	Zip            string         `json:"zip"`
	Latitude       string         `json:"lat"`
	Longitude      string         `json:"lon"`
	NearbyStations NearbyStations `json:"nearby_weather_stations"`
}

// NearbyStations contains info about nearby weather stations
type NearbyStations struct {
	Airport AirportStations `json:"airport"`
	PWS     PWStations      `json:"pws"`
}

// AirportStations contains a slice of AirportStation objects
type AirportStations struct {
	Station []AirportStation `json:"station"`
}

// AirportStation represents a weather station at an airport
type AirportStation struct {
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	ICAO      string `json:"icao"`
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
}

// PWStations contains a slice of PWStation objects
type PWStations struct {
	Station []PWStation `json:"station"`
}

// PWStation represents an unofficial, local weather station
type PWStation struct {
	ID           string  `json:"id"`
	Neighborhood string  `json:"neighborhood"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lon"`
	Distance     int32   `json:"distance_mi"`
}

// Conditions encapsulates WU's conditions API response object
type Conditions struct {
	CurrentObservation CurrentObservation `json:"current_observation"`
}

// CurrentObservation represents the weather conditions right now for a given station
type CurrentObservation struct {
	StationID    string
	Weather      string  `json:"weather"`
	Temperature  float64 `json:"temp_f"`
	HumidityStr  string  `json:"relative_humidity"`
	Humidity     float64
	Dewpoint     float64 `json:"dewpoint_f"`
	WindChillStr string  `json:"windchill_f"`
	WindChill    float64
	HeatIndexStr string `json:"heat_index_f"`
	HeatIndex    float64
	WindDir      float64 `json:"wind_degrees"`
	WindSpeed    float64 `json:"wind_mph"`
	WindGust     float64 `json:"wind_gust_mph"`
	BarometerStr string  `json:"pressure_mb"`
	Barometer    float64
	RainTodayStr string `json:"precip_today_in"`
	RainToday    float64
	Rain1HourStr string `json:"precip_1hr_in"`
	Rain1Hour    float64
}

func (w *WeatherBar) getCurrentConditionsFromWU(icao string, pws string) error {
	var cond Conditions
	var wuURL string

	if icao != "" {
		cond.CurrentObservation.StationID = icao
		wuURL = wgAPIBaseURL + w.cfg.Weather.WUAPIKey + "/conditions/q/icao:" + icao + ".json"

		if *w.debug {
			log.Println("Fetching conditions for ICAO station", icao, "from Weather Underground...")
		}
	} else if pws != "" {
		cond.CurrentObservation.StationID = pws
		wuURL = wgAPIBaseURL + w.cfg.Weather.WUAPIKey + "/conditions/q/pws:" + pws + ".json"

		if *w.debug {
			log.Println("Fetching conditions for PWS station", pws, "from Weather Underground...")
		}
	} else {
		return fmt.Errorf("Must provide either an ICAO station ID or a PWS ID")
	}

	// Fetch the conditions from WU over HTTPS
	var c = &http.Client{Timeout: 10 * time.Second}
	r, err := c.Get(wuURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Decode the JSON we recieved
	err = json.NewDecoder(r.Body).Decode(&cond)
	if err != nil {
		return err
	}

	// WU sends some numeric conditions as strings, so we'll parse them into the proper types
	// and check for obvious errors.
	cond, err = parseConditions(cond)
	if err != nil {
		log.Println("error parsing current conditions:", err)
	}

	w.wxObsChan <- cond.CurrentObservation
	return nil
}

func parseConditions(raw Conditions) (parsed Conditions, err error) {
	// Set all of the parsed data to be the raw data.  We'll fill in the parsed
	// data on top of the raw data as we go along.
	parsed = raw

	parsed.CurrentObservation.Barometer, err = strconv.ParseFloat(raw.CurrentObservation.BarometerStr, 64)
	if err != nil {
		return raw, err
	}

	if raw.CurrentObservation.HeatIndexStr == "NA" {
		parsed.CurrentObservation.HeatIndex = 0
	} else {
		parsed.CurrentObservation.HeatIndex, err = strconv.ParseFloat(raw.CurrentObservation.HeatIndexStr, 64)
		if err != nil {
			return raw, err
		}
	}

	if raw.CurrentObservation.WindChillStr == "NA" {
		parsed.CurrentObservation.WindChill = 0
	} else {
		parsed.CurrentObservation.WindChill, err = strconv.ParseFloat(raw.CurrentObservation.WindChillStr, 64)
		if err != nil {
			return raw, err
		}
	}

	parsed.CurrentObservation.Humidity, err = strconv.ParseFloat(raw.CurrentObservation.HumidityStr[:(len(raw.CurrentObservation.HumidityStr)-1)], 64)
	if err != nil {
		return raw, err
	}

	if raw.CurrentObservation.Rain1HourStr[0] == '-' {
		parsed.CurrentObservation.Rain1Hour = 0
	} else {
		parsed.CurrentObservation.Rain1Hour, err = strconv.ParseFloat(raw.CurrentObservation.Rain1HourStr, 64)
		if err != nil {
			return raw, err
		}
	}

	if raw.CurrentObservation.RainTodayStr[0] == '-' {
		parsed.CurrentObservation.RainToday = 0
	} else {
		parsed.CurrentObservation.RainToday, err = strconv.ParseFloat(raw.CurrentObservation.RainTodayStr, 64)
		if err != nil {
			return raw, err
		}
	}

	return parsed, nil
}
