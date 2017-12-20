package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (w *WeatherBar) getLocationFromFreeGEOIP() (err error) {
	var c = &http.Client{Timeout: 10 * time.Second}
	r, err := c.Get(freeGeoIPURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	w.locMutex.Lock()
	err = json.NewDecoder(r.Body).Decode(&w.loc)
	w.locMutex.Unlock()
	if err != nil {
		return err
	}

	w.locMutex.RLock()
	w.pointMutex.Lock()
	w.point.Latitude = w.loc.Latitude
	w.point.Longitude = w.loc.Longitude
	w.pointMutex.Unlock()
	w.locMutex.RUnlock()

	return nil
}

// Not currently used.  We're using our NOAA database and Haversine distance to
// find the nearest station instead.
func (w *WeatherBar) findNearestICAOStationFromWU() (icao string, err error) {
	var geo GeoLookup

	w.pointMutex.RLock()
	lat := strconv.FormatFloat(w.point.Latitude, 'f', 6, 64)
	lon := strconv.FormatFloat(w.point.Longitude, 'f', 6, 64)
	w.pointMutex.RUnlock()

	wuURL := wgAPIBaseURL + w.cfg.Weather.WUAPIKey + "/geolookup/q/" + lat + "," + lon + ".json"

	var c = &http.Client{Timeout: 10 * time.Second}
	r, err := c.Get(wuURL)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	// w.locMutex.Lock()
	err = json.NewDecoder(r.Body).Decode(&geo)
	// w.locMutex.Unlock()
	if err != nil {
		return "", err
	}

	return geo.Location.NearbyStations.Airport.Station[0].ICAO, nil
}
