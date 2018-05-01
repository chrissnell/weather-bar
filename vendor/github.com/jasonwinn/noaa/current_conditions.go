package noaa

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// The Base URL for NOAA's Current Conditions
// We append to this a Station Id to get the full URL.
const (
	stationConditionsUrl = "http://w1.weather.gov/xml/current_obs/"
	conditionTime        = "Mon, 02 Jan 2006 15:04:05 -0700"
)

// Current Condition represents the Current Conditions
// from NOAA at a given Station.
type CurrentCondition struct {
	XMLName xml.Name `xml:"current_observation"`

	StationId             string  `xml:"station_id"`
	Latitude              float64 `xml:"latitude"`
	Longitude             float64 `xml:"longitude"`
	StringObservationTime string  `xml:"observation_time_rfc822"`
	ObservationTime       time.Time
	TemperatureF          float64 `xml:"temp_f"`
	TemperatureC          float64 `xml:"temp_c"`
	WindDirection         string  `xml:"wind_dir"`
	WindDegrees           float64 `xml:"wind_degrees"`
	WindMph               float64 `xml:"wind_mph"`
	WindGustMph           float64 `xml:"wind_guest_mph"`
	WindKt                float64 `xml:"wind_kt"`
	PressureMB            float64 `xml:"pressure_mb"`
}

// Retrieve the Current Conditions for a Station.
func (station *Station) CurrentConditions() *CurrentCondition {
	c := &CurrentCondition{}

	resp, err := http.Get(stationConditionsUrl + station.Id + ".xml")

	if err != nil {
		return c
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// TODO - (Hack) Convert this to be a proper UTF-8 document.
	// Mahonia package looks interesting but word is there will be an official
	// encoding/decoding package coming soon. We will refactor this when that happens.
	sBody := strings.Replace(string(body), "ISO-8859-1", "UTF-8", -1)
	xml.Unmarshal([]byte(sBody), &c)

	// Manually parse the time
	// encoding/xml doesn't properly encode to a time.Time type
	c.ObservationTime, _ = time.Parse(conditionTime, c.StringObservationTime)
	fmt.Println(c.ObservationTime)

	return c
}
