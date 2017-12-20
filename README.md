![](http://island.nu/github/weather-bar/weather-bar.gif)

# weather-bar
*weather-bar* pulls weather reports from NOAA or Weather Underground and displays them in a desktop bar, like [polybar](https://github.com/jaagr/polybar) or lemonbar.

# How to use

1. Compile weather-bar and put the binary in your `$PATH`:  `go get -u github.com/chrissnell/weather-bar`
2. Configure a [gopherwx](https://github.com/chrissnell/gopherwx) server, or use someone else's.  Make sure you specify a gRPC server in the config.yaml.  If this is for a laptop that will leave your network, expose gopherwx's gRPC port to the Internet via your router.
3. Create your weather-bar configuration file in `${HOME}/.config/weather-bar/config` using the [example](https://github.com/chrissnell/weather-bar/blob/master/example/config) from this repo.
4. Choose one of the options below, depending on your bar.
## Polybar
Add a new module to your Polybar config that uses Polybar's script module to run weather-bar in tail fashion.  See the config [snippet](https://github.com/chrissnell/weather-bar/blob/master/example/polybar-config) in this repo for an example.
## Lemonbar
Simply pipe the output of weather-bar to lemonbar:   `weather-bar | lemonbar`.  I recommend the [patched version](https://github.com/krypt-n/bar) that supports Xft fonts so that you can have some sweet icons.

# Fonts
weather-bar looks best with the Font Awesome icons from [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts).  I used the "Sauce Code Pro" for the screenshot above.
