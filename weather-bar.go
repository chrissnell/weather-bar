package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/jasonwinn/noaa"
)

const freeGeoIPURL = "https://freegeoip.net/json/"

// By default, we poll our geolocation once a day.  If the computer is put into sleep
// mode and awaken, the geolocation will be re-polled automatically regardless of this
// interval.
const geoUpdateInterval = 24 * time.Hour

// Use a 1-hour update interval for NOAA weather.  NOAA updates their conditions hourly,
// at 45 minutes after the hour.  There is no point in fetching more than once an hour.
// TO DO: add support for Weather Underground API.
const noaaUpdateInterval = 1 * time.Hour

// WeatherBar holds our state and useful channels
type WeatherBar struct {
	cfg                 *Config
	loc                 GeoLocation
	locMutex            sync.RWMutex
	prevLoc             GeoLocation
	point               noaa.Point
	pointMutex          sync.RWMutex
	station             *noaa.Station
	stationMutex        sync.RWMutex
	wxObsChan           chan CurrentObservation
	sleepTickerChan     <-chan time.Time
	wxUpdateTickerChan  <-chan time.Time
	wxUpdateChan        chan struct{}
	geoUpdateTickerChan <-chan time.Time
	geoUpdateChan       chan struct{}
	debug               *bool
}

// WeatherObservation holds our current weather observation
type WeatherObservation struct {
	StationID   string
	Temperature float64
	BarometerMb float64
	WindSpeed   float64
	WindDir     float64
}

// GeoLocation holds information returned from the freegeoip.net
// service API
type GeoLocation struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	ZIP         string  `json:"zip_code"`
	Timezone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

func main() {

	w := new(WeatherBar)

	// Get our current UID, used to locate our config file
	uid, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	cfgFile := flag.String("config", uid.HomeDir+"/.config/noaa-weather-bar/config", "Path to noaa-weather-bar config file (default: $HOME/.config/noaa-weather-bar/config)")
	w.debug = flag.Bool("debug", false, "Turn on debugging output")
	flag.Parse()

	// Read our server configuration
	filename, _ := filepath.Abs(*cfgFile)
	w.cfg, err = NewConfig(filename)
	if err != nil {
		log.Fatalln("Error reading config file.  Did you pass the -config flag?  Run with -h for help.\n", err)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func(cancel context.CancelFunc) {
		// If we get a SIGINT or SIGTERM, cancel the context and unblock 'done'
		// to trigger a program shutdown
		<-sigs
		cancel()
		close(done)
	}(cancel)

	w.wxUpdateChan = make(chan struct{}, 1)
	w.geoUpdateChan = make(chan struct{}, 1)
	w.wxObsChan = make(chan CurrentObservation, 1)

	w.wxUpdateTickerChan = time.NewTicker(noaaUpdateInterval).C
	w.geoUpdateTickerChan = time.NewTicker(geoUpdateInterval).C

	go w.weatherWatcher(ctx)
	go w.locationWatcher(ctx)
	go w.sleepDetector(ctx)
	go w.weatherReporter(ctx)

	// Wait for 'done' to unblock before terminating
	<-done
}

func (w *WeatherBar) weatherReporter(ctx context.Context) {
	var output string
	var cardIndex int

	cardDirections := []string{"  N", "NNE", " NE", "ENE",
		"  E", "ESE", " SE", "SSE",
		"  S", "SSW", " SW", "WSW",
		"  W", "WNW", " NW", "NNW"}

	regTempF := regexp.MustCompile("%temperature-fahrenheit%")
	regTempC := regexp.MustCompile("%temperature-celcius%")
	regBar := regexp.MustCompile("%barometer%")
	regWindS := regexp.MustCompile("%wind-speed%")
	regWindD := regexp.MustCompile("%wind-direction%")
	regWindC := regexp.MustCompile("%wind-cardinal%")
	regStationID := regexp.MustCompile("%station-id%")

	output = w.cfg.Format.WxFormat

	for {
		select {
		case obs := <-w.wxObsChan:
			cardIndex = int((float32(obs.WindDir) + float32(11.25)) / float32(22.5))
			cardDirection := cardDirections[cardIndex%16]

			tempC := (obs.Temperature - 32) * (5 / 9)

			output = regTempF.ReplaceAllLiteralString(output, fmt.Sprintf("%.1f", obs.Temperature))
			output = regTempC.ReplaceAllLiteralString(output, fmt.Sprintf("%.1f", tempC))
			output = regBar.ReplaceAllLiteralString(output, fmt.Sprintf("%.2f", obs.Barometer))
			output = regWindS.ReplaceAllLiteralString(output, fmt.Sprintf("%v", obs.WindSpeed))
			output = regWindD.ReplaceAllLiteralString(output, fmt.Sprintf("%v", obs.WindDir))
			output = regWindC.ReplaceAllLiteralString(output, cardDirection)
			output = regStationID.ReplaceAllLiteralString(output, fmt.Sprintf("%v", obs.StationID))

			fmt.Println(output)

		case <-ctx.Done():
			log.Println("Termination request recieved.  Cancelling weather watcher.")
			return

		}
	}
}

func (w *WeatherBar) weatherWatcher(ctx context.Context) {
	var err error

	for {
		select {
		case <-w.wxUpdateTickerChan:
			// We got a tick, so just trigger the wx update channel
			w.wxUpdateChan <- struct{}{}
		case <-w.wxUpdateChan:

			// If our station ID is not yet set, that's probably because the geolocation hasn't
			// finished.  Sleep until it has.
			for {
				w.stationMutex.RLock()
				if w.station.Id == "" {
					w.stationMutex.RUnlock()
					if *w.debug {
						log.Println("Station ID not yet determined.  Sleeping 1s")
					}
					time.Sleep(time.Second)
				} else {
					w.stationMutex.RUnlock()
					break
				}
			}

			w.stationMutex.RLock()
			// If a Weather Underground API key has been set in the config, use that
			// service to fetch the weather conditions.  Otherwise, use NOAA.
			if w.cfg.Weather.WUAPIKey != "" {

				// WU supports two types of stations: ICAO (official government-run stations)
				// and PWS (personal weather stations, typically run by individuals, businesses, etc.)
				// If our station ID is 4 bytes long, it's almost certainly an ICAO station, so
				// we pass it to our getter as such.  Otherwise, we pass the station ID as a PWS
				// ID.
				if len(w.station.Id) == 4 {
					err = w.getCurrentConditionsFromWU(w.station.Id, "")
				} else {
					err = w.getCurrentConditionsFromWU("", w.station.Id)
				}
				if err != nil {
					log.Println(err)
					continue
				}
			} else {
				err := w.getCurrentConditionsFromNOAA(w.station.Id)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			w.stationMutex.RUnlock()

		case <-ctx.Done():
			log.Println("Termination request recieved.  Cancelling weather watcher.")
			return

		}
	}

}

func (w *WeatherBar) locationWatcher(ctx context.Context) {
	if w.cfg.Weather.Station != "" {
		// We were given a NOAA station ID in our config file so we will use that and
		// forego any further geolocation activities by exiting this goroutine
		if *w.debug {
			log.Printf("Weather station is hardcoded (%v).  Disabling geolocation.\n", w.cfg.Weather.Station)
		}
		w.stationMutex.Lock()
		w.station = &noaa.Station{Id: w.cfg.Weather.Station}
		w.stationMutex.Unlock()

		// Since we're just starting up, force a weather update.
		w.wxUpdateChan <- struct{}{}

		return
	}

	// Force a geolocation update.  We'll need a starting point in order to monitor
	// for location changes.
	err := w.getLocationFromFreeGEOIP()
	if err != nil {
		log.Fatalln("could not get location:", err)
	}

	// Update our previous location with our current location
	w.locMutex.RLock()
	w.prevLoc = w.loc
	if *w.debug {
		log.Printf("LOCATION: %+v\n", w.loc)
	}
	w.locMutex.RUnlock()

	// Find our nearest weather station and update our station object
	w.pointMutex.RLock()
	w.stationMutex.Lock()

	w.station = w.point.NearestStation()
	if *w.debug {
		log.Println("STATION:", w.station.Id)
	}

	w.stationMutex.Unlock()
	w.pointMutex.RUnlock()
	// Since we're just starting up, force a weather update.
	w.wxUpdateChan <- struct{}{}

	for {
		select {
		case <-w.geoUpdateTickerChan:
			// We got a tick, so just trigger the update channel
			w.geoUpdateChan <- struct{}{}
		case <-w.geoUpdateChan:
			// Check to see if the user hard-coded a lat/lon in the config file.
			// If the user provided a lat/lon, we don't need to geolocate.
			if (w.point.Latitude == 0) || (w.point.Longitude == 0) {
				// Our geolocation update timer has ticked, so we'll run a geolocation
				// update and see if the location has changed.
				// Kick off a geolocation update...
				err = w.getLocationFromFreeGEOIP()
				if err != nil {
					// We failed to geolocate.  Sleep 15 seconds and give it another shot.
					log.Println("error fetching location:", err)
					time.Sleep(15 * time.Second)
					err = w.getLocationFromFreeGEOIP()
					if err != nil {
						// Geolcoation failed a second time so we'll just wait for the next
						// scheduled geolocation update timer to fire
						log.Println("error fetching location:", err)
						continue
					}
				}
			} else {
				if *w.debug {
					log.Printf("User provided location (%v/%v). Skipping geolocation\n", w.cfg.Weather.Latitude, w.cfg.Weather.Longitude)
				}
			}

			// Find our nearest weather station and update our station object
			w.pointMutex.RLock()
			w.stationMutex.Lock()
			station := w.point.NearestStation()
			w.stationMutex.Unlock()
			w.pointMutex.RUnlock()
			if *w.debug {
				log.Println("STATION:", station.Id)
			}

			w.locMutex.RLock()
			if (w.loc.Latitude != w.prevLoc.Latitude) || (w.loc.Longitude != w.prevLoc.Longitude) {
				// We've moved, so let's kick off a weather update.
				w.wxUpdateChan <- struct{}{}
			}

			// Set our previous location to our current location
			w.prevLoc = w.loc
			if *w.debug {
				log.Printf("LOCATION: %+v\n", w.loc)
			}
			w.locMutex.RUnlock()

		case <-ctx.Done():
			log.Println("Termination request recieved.  Cancelling weather watcher.")
			return

		}
	}
}

func (w *WeatherBar) sleepDetector(ctx context.Context) {
	w.sleepTickerChan = time.NewTicker(time.Second * 1).C
	prevTime := time.Now()

	for {
		select {
		case t := <-w.sleepTickerChan:
			t = t.Round(0)
			prevTime = prevTime.Round(0)
			dur := t.Sub(prevTime)
			if dur > 30*time.Second {
				log.Println("WAKEUP DETECTED!!!")
				// We've been sleeping so check our location and update weather if necessary
				w.geoUpdateChan <- struct{}{}
			}
			prevTime = t

		case <-ctx.Done():
			log.Println("Termination request recieved.  Cancelling weather watcher.")
			return

		}
	}
}
