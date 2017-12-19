package main

import (
	"io/ioutil"

	"github.com/go-ini/ini"
)

// Config is the base configuraiton object
type Config struct {
	Weather WeatherConfig
	Format  FormatConfig
}

// WeatherConfig holds configuration related to our local weather station
type WeatherConfig struct {
	Latitude  string `ini:"latitude"`
	Longitude string `ini:"longitude"`
	Station   string `ini:"station"`
}

// FormatConfig holds our output formatting configuration
type FormatConfig struct {
	WxFormat string `ini:"weather-format"`
}

// NewConfig creates an new config object from the given filename.
func NewConfig(filename string) (*Config, error) {
	c := new(Config)
	cfgFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Config{}, err
	}

	cfg, err := ini.Load(cfgFile)
	if err != nil {
		return &Config{}, err
	}

	err = cfg.Section("weather").MapTo(&c.Weather)
	if err != nil {
		return &Config{}, err
	}
	err = cfg.Section("format").MapTo(&c.Format)
	if err != nil {
		return &Config{}, err
	}

	return c, nil
}
