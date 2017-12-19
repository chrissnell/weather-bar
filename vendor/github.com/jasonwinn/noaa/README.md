NOAA Weather Forecasts & Current Conditions
=============================================

## What It Does
* Finds the nearest weather station with a given longitude and latitude 
* Gives the current weather conditions for a weather station
* Gives a Weather Forecast, given a longitude and latitude (finds the nearest weather station)

## Install

* go get "github.com/jasonwinn/noaa"
* import "github.com/jasonwinn/noaa"


## Examples

#### Current Conditions
```
point := noaa.Point{Latitude: 71.290556, Longitude: -156.788611} 
station := point.NearestStation()
conditions := station.CurrentConditions()

// Current Conditions Example Output:
conditions.StationId        //    "PABR"
conditions.Latitude         //    71.29
conditions.Longitude        //    156.77 
conditions.TemperatureF     //    50 
conditions.TemperatureC     //    10
conditions.WindDirection    //    "Southwest" 
conditions.WindDegrees      //    230
conditions.WindMph          //    4.6
conditions.WindGustMph      //    0
conditions.WindKt           //    4 
conditions.PressureMB       //    1015.8
```


#### Forecast
```
point := noaa.Point{Latitude: 71.290556, Longitude: -156.788611} 
forecast := point.Forecast(5) // Return 5 forecast days 

// Forecast Example Output:
forecast.Length        // 5 (Number of days of forecast)
forecast.Point         // xxx, xxx (Actual point of the forecast)
forecast.ForecastDays  // []ForecastDay 

for _, forecastDay := range forecast.ForecastDays {
  fmt.Println(forecastDay)
}

// forecastDay Example Output:
forecastDay.MaxTemperature            // 5  // fahrenheit 
forecastDay.MinTemperature            // 36 // fahrenheit
forecastDay.SummaryDay                // map
forecastDay.SummaryDay["summary"]     // "Partly Cloudy"
forecastDay.SummaryDay["icon"]        // "http://www.nws.noaa.gov/weather/images/fcicons/nsct.jpg"
forecastDay.SummaryNight              // map
forecastDay.SummaryNight["summary"]   // "Partly Cloudy"
forecastDay.SummaryNight["icon"]      // "http://www.nws.noaa.gov/weather/images/fcicons/nsct.jpg"
forecastDay.ConditionsDay             // []Condition // Array of Conditions
forecastDay.ConditionsNight           // []Condition // Array of Conditions
forecastDay.PrecipitationChanceDay    // 3
forecastDay.PrecipitationChanceNight  // 4
```

#### ForecastDay.ConditionsDay / ForecastDay.ConditionsNight
```
point := noaa.Point{Latitude: 71.290556, Longitude: -156.788611} 
forecast := point.Forecast(5) // Return 5 forecast days 
for _, forecastDay := range forecast.ForecastDays {
  for _, condition := range forecastDay.ConditionsDay {
    fmt.Println(condition)
  }
}

// Condition Example Output:
condition.Coverage        // scattered
condition.Intensity       // light
condition.WeatherType     // rain showers
condition.Qualifier       // none 
condition.Additive        // and
```

## Refresh NOAA Station List
```
station := point.NearestStation()
```
When you fetch the nearest Station, a list of all available NOAA weather stations is pulled from NOAA. This is an expensive call, the resulting file from NOAA is approximately 1.2 megabytes. Therefore it is generated once and stored in a local xml file. 

If you believe your weather station list is out of date, you can manually refresh it by calling noaa.RefreshNOAAStationList(). This will delete the existing xml file and replace it with the latest available data from NOAA.  


## License
MIT License



