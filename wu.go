package main

// GeoLookup encapsulates WU's response object
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
	ID           string `json:"id"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Latitude     string `json:"lat"`
	Longitude    string `json:"lon"`
	Distance     int32  `json:"distance_mi"`
}
