package noaa

import (
	"encoding/xml"
)

type XmlResultPoint struct {
	XMLName   xml.Name `xml:"point"`
	Latitude  float64  `xml:"latitude,attr"`
	Longitude float64  `xml:"longitude,attr"`
}
type XmlResultTemperature struct {
	XMLName xml.Name  `xml:"temperature"`
	Values  []float64 `xml:"value"`
}
type XmlResultPrecipitationChance struct {
	XmlName xml.Name  `xml:"probability-of-precipitation"`
	Values  []float64 `xml:"value"`
}
type XmlResultWeatherDetails struct {
	XmlName    xml.Name `xml:"weather"`
	Conditions []struct {
		Summary string `xml:"weather-summary,attr"`
		Values  []struct {
			Coverage    string `xml:"coverage,attr"`
			Intensity   string `xml:"intensity,attr"`
			WeatherType string `xml:"weather-type,attr"`
			Qualifier   string `xml:"qualifier,attr"`
			Additive    string `xml:"additive,attr"`
		} `xml:"value"`
	} `xml:"weather-conditions"`
}
type XmlResultConditionIcons struct {
	XmlName xml.Name `xml:"conditions-icon"`
	Links   []string `xml:"icon-link"`
}

// Day -> <layout-key>k-p24h-n7-1</layout-key>
// Night -> <layout-key>k-p24h-n7-2</layout-key>
// Full -> <layout-key>k-p12h-n14-3</layout-key>
type XmlResultPeriod struct {
	XmlName    xml.Name `xml:"time-layout"`
	StartTimes []struct {
		Time       string `xml:",innerxml"`
		PeriodName string `xml:"period-name,attr"`
	} `xml:"start-valid-time"`
	EndTimes []struct {
		Time string `xml:",innerxml"`
	} `xml:"end-valid-time"`
}
