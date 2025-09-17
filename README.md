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

[[gesture_bindings]]
gesture = "swipe_left"
action  = "set_color"
[gesture_bindings.selector]
type = "serial"
value = "d073d5000000"
[gesture_bindings.hsbk]
hue = 240
saturation = 100
brightness = 100

[[gesture_bindings]]
gesture = "swipe_right"
action  = "set_color"
[gesture_bindings.selector]
type = "serial"
value = "d073d5000000"
[gesture_bindings.hsbk]
hue = 0
saturation = 100
brightness = 100

[[finger_bindings]]
pattern = [1,1,1,1,1]
action  = "power_on"
[finger_bindings.selector]
type = "all"

[[finger_bindings]]
pattern = [0,0,0,0,0]
action  = "power_off"
[finger_bindings.selector]
type = "all"
```

### Sections

- [general]: Global settings.
- [logging]: Controls the logging level and output file. Leave file empty for console output.
- [[gesture_bindings]]: Map gestures detected by Fingertrack to actions on your devices.
- [[finger_bindings]]: Map finger patterns to actions.

### Selector

Each gesture or finger binding should include a selector to target a specific device or group, and optional parameters like hsbk for color control.
The accepted selector are:

- all -> target all the discovered devices
- label -> target a single device by label
- group -> target devices with the given group label
- location -> target devices with the given location label
- serial -> target a device with the given serial (e.g., d073d5000000)

### Action

Supported actions are:

- power_on
- power_off
- set_color -> requires at least one of the HSBK (Hue, Saturation, Brightness, Kelvin) to be set

## License

MIT
