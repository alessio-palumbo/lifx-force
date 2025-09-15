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

## License

MIT
