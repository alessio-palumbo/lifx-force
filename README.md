# lifx-force

**lifx-force** is a cross-platform Go application for controlling [LIFX smart lights](https://www.lifx.com/) via the LAN protocol.
It embeds [fingertrack](https://github.com/alessio-palumbo/fingertrack) as a bundled runtime dependency, enabling gesture-based control.

## Features (WIP)

- 🚀 Cross-platform (Linux, macOS, Windows)
- 💡 Discover and control LIFX bulbs on your local network
- ✋ Gesture-based control with [fingertrack]
- 🔒 Runs fully local (no cloud dependency)

## Installation

Download the appropriate archive for your platform from the [Releases](../../releases) page.
Each archive contains:

- `lifx-force` binary with embedded `fingertrack` runtime
- `README.md`
- `LICENSE`
- `VERSION`

Example:

```bash
wget https://github.com/alessio-palumbo/lifx-force/releases/download/v0.1.0/lifx-force-linux-amd64.zip
unzip lifx-force-linux-amd64.zip
cd lifx-force-linux-amd64
./lifx-force
```

## Configuration

On the first run, lifx-force automatically creates a default configuration file at:
~/.lifx-force/config.toml

This file contains basic default values which you can edit to customize the behavior of the app.

Example Configuration

```yaml
[general]
transition_ms = 1          # defines the speed of the light transition defined by the action (min 1ms)

[logging]
level = "info"             # one of: debug, info, warn, error
file  = "lifx-force.log"   # leave empty for stdout

[[bindings]]
gesture = "swipe_left"
action  = "set_color"
[bindings.selector]
type = "serial"
value = "d073d5000000"
[bindings.hsbk]
hue = 240
saturation = 100
brightness = 100

[[bindings]]
gesture = "swipe_right"
action  = "set_color"
[bindings.selector]
type = "serial"
value = "d073d5000000"
[bindings.hsbk]
hue = 0
saturation = 100
brightness = 100

[[bindings]]
pattern = [1,1,1,1,1]
action  = "power_on"
[bindings.selector]
type = "all"

[[bindings]]
pattern = [0,0,0,0,0]
action  = "power_off"
[bindings.selector]
type = "all"
```

### Sections

- [general]: Global settings.
- [logging]: Controls the logging level and output file. Leave file empty for console output.
- [[bindings]]: Map gestures or finger patterns detected by Fingertrack to actions on your devices.

### Gestures

Supported gestures are:

- swipe_left
- swipe_right
- swipe_up
- swipe_down

Supported compound gestures are:

- expand -> triggered when hands move apart from each other on the horizontal plane
- contract -> triggered when hands move closer to each other on the horizontal plane
- push_down -> triggered when both hands push downward
- pull_up -> triggered when both hands rise upward

### Patterns

A pattern describes the state of fingers in a hand, from thumb to pinky, left to right,
with 1 meaning extended and 0 meaning retracted.
E.g.

- [0,0,0,0,0] -> fist
- [1,1,1,1,1] -> open hand

### Action

Supported actions are:

- power_on
- power_off
- set_color -> requires at least one of the HSBK (Hue, Saturation, Brightness, Kelvin) to be set

### Selector

Each binding should include a selector to target a specific device or group, and optional parameters like hsbk for color control.
The accepted selector are:

- all -> target all the discovered devices
- label -> target a single device by label
- group -> target devices with the given group label
- location -> target devices with the given location label
- serial -> target a device with the given serial (e.g., d073d5000000)

## License

MIT
