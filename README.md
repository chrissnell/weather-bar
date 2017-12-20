![](http://island.nu/github/weather-bar/weather-bar.gif)

# weather-bar
**weather-bar** pulls weather reports from NOAA or Weather Underground and displays them in a desktop bar, like [polybar](https://github.com/jaagr/polybar) or lemonbar.  weather-bar can use **geolocation** to find the nearest NOAA weather station and will automatically update your location if you put your laptop to sleep and wake it up in a new location.  Much like polybar and lemonbar, weather-bar is **customizable**.  Using tokens in the config file, you can tweak the display to your liking.

# How to use

1. Download a [release](https://github.com/chrissnell/weather-bar/releases) or compile your own binary and put the binary in your `$PATH` (e.g. `/usr/local/bin/weather-bar`)
2. Using the [example config file found here](https://github.com/chrissnell/weather-bar/blob/master/example/config), create a config file in your `$XDG_CONFIG_HOME` directory (default config file is `${HOME}/.config/weather-bar/config`) and edit if you desire.  
3. Edit your bar config, as described below. 

## Polybar
Add a new module to your Polybar config that uses Polybar's script module to run weather-bar in tail fashion.  See the config [snippet](https://github.com/chrissnell/weather-bar/blob/master/example/polybar-config) in this repo for an example.
## Lemonbar
Simply pipe the output of weather-bar to lemonbar:   `weather-bar | lemonbar`.  I recommend the [patched version](https://github.com/krypt-n/bar) that supports Xft fonts so that you can have some sweet icons.

## Weather Underground support
By default, weather-bar fetches weather conditions from [NOAA](http://www.weather.gov/) but if you [sign up for a free API key](https://www.wunderground.com/api), weather-bar can fetch metrics from the Weather Underground, which gives you much more frequent weather updates (5 minutes vs. 1 hour for NOAA) and the option to pull weather from the large network of personal weather stations (PWS) that send data to WU.

## Geolocation
By default, weather-bar uses [freegeoip.net](https://freegeoip.net) to geolocate your computer and computes the [Haversine distance](https://en.wikipedia.org/wiki/Haversine_formula) to nearby NOAA weather stations to determine the closest one.  If  the geolocation is inaccurate or if you don't wish to geolocate, you can specify a specific NOAA station by its [ICAO code](https://en.wikipedia.org/wiki/ICAO_airport_code) or a Weather Underground Personal Weather Station (PWS) by its ID.  You can also specify a particular latitude and longitude and weather-bar will find the nearest NOAA station automatically.

## Fonts
weather-bar looks best with the Font Awesome icons from [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts).  I used the "Sauce Code Pro" for the screenshot above.
