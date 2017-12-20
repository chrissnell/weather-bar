package main

import (
	"fmt"
	"log"
)

func (w *WeatherBar) getCurrentConditionsFromNOAA(icao string) error {
	if *w.debug {
		log.Println("Fetching conditions for station", icao, "from NOAA...")
	}
	// Fetch current conditions
	w.stationMutex.RLock()
	conditions := w.station.CurrentConditions()
	w.stationMutex.RUnlock()

	if *w.debug {
		log.Printf("CONDITIONS: %+v\n", conditions)
	}

	// If we didn't get back a StationID, something went wrong
	// with weather fetching so don't bother sending a new observation.
	if conditions.StationId == "" {
		return fmt.Errorf("unable to fetch observation for %v", w.station.Id)
	}

	obs := CurrentObservation{
		StationID:   conditions.StationId,
		Temperature: conditions.TemperatureF,
		Barometer:   conditions.PressureMB,
		WindSpeed:   conditions.WindMph,
		WindDir:     conditions.WindDegrees,
	}

	w.wxObsChan <- obs
	return nil
}
