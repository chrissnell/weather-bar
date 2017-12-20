![](http://island.nu/github/weather-bar/weather-bar.gif)

# weather-bar
*weather-bar* pulls weather reports from NOAA or Weather Underground and displays them in a desktop bar, like [polybar](https://github.com/jaagr/polybar) or lemonbar.

# How to use

1. Download a [release](https://github.com/chrissnell/weather-bar/releases) or compile your own binary and put the binary in your `$PATH` (e.g. `/usr/local/bin/weather-bar`)
2. Using the [example config file found here](https://github.com/chrissnell/weather-bar/blob/master/example/config), create a config file in your `$XDG_CONFIG_HOME` directory (default config file is `${HOME}/.config/weather-bar/config`) and edit if you desire.  
3. Edit your bar config, as described below. 

## Polybar
Add a new module to your Polybar config that uses Polybar's script module to run weather-bar in tail fashion.  See the config [snippet](https://github.com/chrissnell/weather-bar/blob/master/example/polybar-config) in this repo for an example.
## Lemonbar
Simply pipe the output of weather-bar to lemonbar:   `weather-bar | lemonbar`.  I recommend the [patched version](https://github.com/krypt-n/bar) that supports Xft fonts so that you can have some sweet icons.

# Getting better and more frequent weather metrics
By default, weather-bar fetches weather conditions from [NOAA](http://www.weather.gov/) but if you [sign up for a free API key](https://www.wunderground.com/api), weather-bar can fetch metrics from the Weather Underground, which gives you much more frequent weather updates (5 minutes vs. 1 hour for NOAA) and the option to pull weather from the large network of personal weather stations (PWS) that send data to WU.

# Fonts
weather-bar looks best with the Font Awesome icons from [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts).  I used the "Sauce Code Pro" for the screenshot above.
