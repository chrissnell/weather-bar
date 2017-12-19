package noaa

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	forecastUrl             = "http://www.weather.gov/forecasts/xml/sample_products/browser_interface/ndfdBrowserClientByDay.php"
	startDay                = "Today"
	startNight              = "Tonight"
	noaaTime                = "2006-01-02T15:04:05-07:00"
	nodePoint               = "point"
	nodeTemperature         = "temperature"
	nodeMaximum             = "maximum"
	nodeMinimum             = "minimum"
	nodePrecipitationChance = "probability-of-precipitation"
	nodeWeather             = "weather"
	nodeConditions          = "conditions-icon"
)

// Weather Options to fetch from NOAA
// NOAA will also return the forecast in summary and detail form
var forecastOptions = []string{
	"maxt",  // maximum temp
	"mint",  // minimum temp
	"pop12", // 12 Hour Probability of Precipitation
}

// Forecast represents the length of a forecast in days,
// the point the forecast is valid for and an array of forecasts per day.
type Forecast struct {
	Length       int
	Point        Point // Return the point of actual forecast
	ForecastDays []ForecastDay
}

// ForecastDay represents the forecast for a given day.
// SummaryDay/Night contains a map representing the "summary" and "image"
// for that time period. "image" is the graphical representation of the forecast
// returned by NOAA.
// ConditionsDay/Night return the details Conditions for a time period. It is an array
// because there can be multiple conditions, e.g. Partly Cloudy and Snowy.
type ForecastDay struct {
	NameDay                  string
	NameNight                string
	StartTime                time.Time
	EndTime                  time.Time
	MaxTemperature           float64
	MinTemperature           float64
	SummaryDay               map[string]string // Party Cloud, image url
	SummaryNight             map[string]string // Party Cloud, image url
	ConditionsDay            []Condition
	ConditionsNight          []Condition
	PrecipitationChanceDay   float64
	PrecipitationChanceNight float64
}

// Detailed Conditions for a Time Period.
type Condition struct {
	Coverage    string
	Intensity   string
	WeatherType string
	Qualifier   string
	Additive    string
}

// Xml Result Collections
var forecastPoint = &XmlResultPoint{}
var maxTemps = &XmlResultTemperature{}
var minTemps = &XmlResultTemperature{}
var precipitationChances = &XmlResultPrecipitationChance{}
var weatherDetails = &XmlResultWeatherDetails{}
var conditionIcons = &XmlResultConditionIcons{}
var periodDay = &XmlResultPeriod{}
var periodFull = &XmlResultPeriod{}

// Returns a Forecast closest to a given Point
func (p *Point) Forecast(daysRequested int) *Forecast {
	url := forecastUrl +
		"?startDate=" +
		"&lat=" + fmt.Sprintf("%f", p.Latitude) +
		"&lon=" + fmt.Sprintf("%f", p.Longitude) +
		"&format=" + "12+hourly" +
		"&numDays=" + fmt.Sprintf("%d", daysRequested)

	// Options to return
	for _, option := range forecastOptions {
		url += "&" + option + "=" + option
	}

	fmt.Println("Fetching: " + url)

	resp, err := http.Get(url)

	if err != nil {
		// return an empty forecast
		return &Forecast{}
	}

	// We are converting the response into a []byte, which makes it a little easier to pass around,
	// especially for our tests which are reading local files which are []byte.
	// This also helps us kill the http quicker, instead of waiting for the XML decoder to finish.
	// The only con is that we have to covert this back to a Reader for Xml.NewDecode
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// Forecast Struct
	forecast := &Forecast{Length: daysRequested, ForecastDays: make([]ForecastDay, daysRequested)}
	forecast.assignForecastData(body)

	return forecast
}

// NOAA returns the forecast data in a variety of formats. This function unmarshals
// the given data and then assigns into ForecastDays and creates a Forecast
func (f *Forecast) assignForecastData(body []byte) {
	// Unmarshal the forecast xml file
	unmarshalForecast(body)

	// Point of actual forecast
	f.Point.Latitude = forecastPoint.Latitude
	f.Point.Longitude = forecastPoint.Longitude

	// Forecasts are in 12 hour periods
	f.Length = len(periodDay.StartTimes)

	// Put the collections into appropriate forecast structs
	f.ForecastDays = make([]ForecastDay, f.Length)

	// NOAA's forecasts can begin with a day forecast or a night forecast.
	// Furthermore, some of the forecast values are in 12 hour groups while others are in
	// 24 hour groups. Depending on whether the first 12 hour forecast is for a day or night
	// we use the dayPosition and nightPosition values below to know where to pull from the 12 hour group values.
	// We advace position twice as fast as i.
	position := 0
	var dayPosition int
	var nightPosition int

	for i := range periodDay.StartTimes {
		d := &f.ForecastDays[i]

		if periodFull.StartTimes[0].PeriodName == startDay {
			dayPosition = position
			nightPosition = position + 1
		} else {
			dayPosition = position + 1
			nightPosition = position
		}

		// Maximum / Minimum Temps (24 Hour Groups)
		if i < len(maxTemps.Values) {
			d.MaxTemperature = maxTemps.Values[i]
		}
		if i < len(minTemps.Values) {
			d.MinTemperature = minTemps.Values[i]
		}

		// Precipitation Chance (12 Hour Groups)
		if dayPosition < len(precipitationChances.Values) {
			d.PrecipitationChanceDay = precipitationChances.Values[dayPosition]
		}
		if nightPosition < len(precipitationChances.Values) {
			d.PrecipitationChanceNight = precipitationChances.Values[nightPosition]
		}

		// Summary (12 Hour Groups)
		d.SummaryDay = make(map[string]string)
		d.SummaryNight = make(map[string]string)

		// Summary Text / Condition Details (12 Hour Groups)
		if dayPosition < len(weatherDetails.Conditions) {
			d.SummaryDay["summary"] = weatherDetails.Conditions[dayPosition].Summary

			d.ConditionsDay = make([]Condition, len(weatherDetails.Conditions[dayPosition].Values))

			for j := range weatherDetails.Conditions[dayPosition].Values {
				result := weatherDetails.Conditions[dayPosition].Values[j]
				c := Condition{
					Coverage:    result.Coverage,
					Intensity:   result.Intensity,
					WeatherType: result.WeatherType,
					Qualifier:   result.Qualifier,
					Additive:    result.Additive,
				}

				d.ConditionsDay[j] = c
			}
		}
		// Summary Text / Condition Details (12 Hour Groups)
		if nightPosition < len(weatherDetails.Conditions) {
			d.SummaryNight["summary"] = weatherDetails.Conditions[nightPosition].Summary

			d.ConditionsNight = make([]Condition, len(weatherDetails.Conditions[nightPosition].Values))
			for j := range weatherDetails.Conditions[nightPosition].Values {
				result := weatherDetails.Conditions[nightPosition].Values[j]
				c := Condition{
					Coverage:    result.Coverage,
					Intensity:   result.Intensity,
					WeatherType: result.WeatherType,
					Qualifier:   result.Qualifier,
					Additive:    result.Additive,
				}

				d.ConditionsNight[j] = c
			}
		}

		// Summary Image (12 Hour Groups)
		if dayPosition < len(conditionIcons.Links) {
			d.SummaryDay["icon"] = conditionIcons.Links[dayPosition]

		}
		if nightPosition < len(conditionIcons.Links) {
			d.SummaryNight["icon"] = conditionIcons.Links[nightPosition]
		}

		// Set the Time Periods (12 Hour Groups)
		if dayPosition < len(periodFull.StartTimes) {
			// Name "Saturday"
			d.NameDay = periodFull.StartTimes[dayPosition].PeriodName

			// Time
			t, _ := time.Parse(noaaTime, periodFull.StartTimes[dayPosition].Time)
			d.StartTime = t
		}
		if nightPosition < len(periodFull.EndTimes) {
			// Name "Saturday Night"
			d.NameNight = periodFull.StartTimes[nightPosition].PeriodName

			// Time
			t, _ := time.Parse(noaaTime, periodFull.EndTimes[nightPosition].Time)
			d.EndTime = t
		}

		position += 2
	}
}

// Searches the XML file for weather data and unmarshals it into XML result collections.
func unmarshalForecast(body []byte) {
	decoder := xml.NewDecoder(bytes.NewReader(body))

	// Loop through the elements
	// Decode into formats we can work with
	summaryLevel := 0
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch s := token.(type) {
		case xml.StartElement:

			// Decode point
			if s.Name.Local == nodePoint {
				decoder.DecodeElement(&forecastPoint, &s)
			}

			// Decode temperature array
			if s.Name.Local == nodeTemperature {
				for _, attr := range s.Attr {
					// Unmarshal maximum temps
					if attr.Value == nodeMaximum {
						decoder.DecodeElement(&maxTemps, &s)
					}

					// Unmarshal minimum temps
					if attr.Value == nodeMinimum {
						decoder.DecodeElement(&minTemps, &s)
					}
				}
			}

			// Decode precipitation chance array
			if s.Name.Local == nodePrecipitationChance {
				decoder.DecodeElement(&precipitationChances, &s)
			}

			// Decode weather summary collections
			if s.Name.Local == nodeWeather {
				decoder.DecodeElement(&weatherDetails, &s)
			}

			// Decode Condition Icon Urls
			if s.Name.Local == nodeConditions {
				decoder.DecodeElement(&conditionIcons, &s)
			}

			if s.Name.Local == "time-layout" {
				summaryLevel++

				// The summary levels all have the same tag names,
				// so we need to do some manual checking here.
				// DayTime
				if summaryLevel == 1 {
					decoder.DecodeElement(&periodDay, &s)
				}
				// Everything (Night + Day || Day + Night)
				if summaryLevel == 3 {
					decoder.DecodeElement(&periodFull, &s)
				}

			}
		}
	}
}
