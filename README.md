# gocrosshair

A lightweight, high-performance crosshair overlay for Linux (X11/XWayland).

## Features

- **Zero runtime dependencies**: Pure Go implementation using X11 protocol directly
- **Extremely low resource usage**: No GTK, Qt, or Python required
- **Click-through**: Mouse events pass through to underlying applications
- **Always on top**: Uses override-redirect to bypass window manager
- **Works on Wayland**: Compatible via XWayland
- **Multi-monitor support**: Choose which monitor to display crosshair on
- **Configurable**: TOML config file with shape, color, size, and position options

## Requirements

- Linux with X11 or XWayland
- Go 1.21+ (for building)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/gocrosshair
cd gocrosshair

# Build optimized binary
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o gocrosshair .

# Run
./gocrosshair
```

### Arch Linux (AUR)

```bash
# Using an AUR helper
yay -S gocrosshair

# Or manually
git clone https://aur.archlinux.org/gocrosshair.git
cd gocrosshair
makepkg -si
```

## Usage

Simply run the binary:

```bash
gocrosshair
```

On first run, a default configuration file is created at `~/.config/gocrosshair/config.toml`.

### Command-Line Options

```
gocrosshair [options]

Options:
  -config string      Path to configuration file (default: ~/.config/gocrosshair/config.toml)
  -list-monitors      List available monitors and exit
  -setup              Run interactive setup wizard to create configuration
  -stop               Stop any running gocrosshair instance
  -version            Show version and exit
  -help               Show help message
```

### Interactive Setup Wizard

Create or modify your configuration interactively:

```bash
gocrosshair -setup
```

The wizard guides you through:
- Crosshair shape selection (cross, dot, circle, cross-dot)
- Color selection with presets or custom hex colors
- Size, thickness, and gap configuration
- Outline options
- Monitor selection
- Position offset from center

Use arrow keys to navigate, Enter to select, and Esc to go back.

### Running in Background

Start the crosshair in the background:

```bash
gocrosshair &
```

Stop it later:

```bash
gocrosshair -stop
```

The application uses a PID file (`$XDG_RUNTIME_DIR/gocrosshair.pid` or `/tmp/gocrosshair.pid`) to track the running instance. Only one instance can run at a time.

### List Available Monitors

```bash
gocrosshair -list-monitors
```

Output:
```
Available monitors:

  [0] DP-1: 1920x1080 at position (0, 0) ← primary
  [1] DP-2: 1920x1080 at position (1920, 0)

Use 'monitor = N' in config to select a monitor by index.
Use 'monitor = -1' to automatically select the primary monitor.
```

## Configuration

The configuration file is located at `~/.config/gocrosshair/config.toml` (or `$XDG_CONFIG_HOME/gocrosshair/config.toml`).

### Full Configuration Example

```toml
[crosshair]
# Shape: "cross", "dot", "circle", "cross-dot"
shape = "cross"

# Color in hex format (#RRGGBB, 0xRRGGBB, or RRGGBB)
color = "#00FF00"

# Size of the crosshair arms in pixels (from center)
size = 10

# Thickness of lines in pixels
thickness = 2

# Gap in center (pixels) - creates hollow cross shape
gap = 0

# Outline settings (set outline_thickness to 0 to disable)
outline_thickness = 0
outline_color = "#000000"

[position]
# Monitor index (0 = first, 1 = second, etc.)
# Use -1 for primary monitor
monitor = 0

# Offset from monitor center (pixels)
# Positive X = right, Positive Y = down
offset_x = 0
offset_y = 0
```

### Shape Examples

#### Classic Cross
```toml
[crosshair]
shape = "cross"
color = "#00FF00"
size = 10
thickness = 2
gap = 0
```

#### CS-Style Crosshair (with gap)
```toml
[crosshair]
shape = "cross"
color = "#00FF00"
size = 8
thickness = 2
gap = 4
```

#### Center Dot
```toml
[crosshair]
shape = "dot"
color = "#FF0000"
size = 4
```

#### Circle
```toml
[crosshair]
shape = "circle"
color = "#FFFFFF"
size = 6
```

#### Cross with Outline
```toml
[crosshair]
shape = "cross"
color = "#FFFFFF"
size = 12
thickness = 2
gap = 4
outline_thickness = 1
outline_color = "#000000"
```

### Common Colors

| Color  | Hex Code  |
|--------|-----------|
| Green  | `#00FF00` |
| Red    | `#FF0000` |
| White  | `#FFFFFF` |
| Cyan   | `#00FFFF` |
| Yellow | `#FFFF00` |
| Pink   | `#FF00FF` |

## Error Handling

If the configuration file is invalid, the application will prompt you:

```
╭─ Configuration Error ─────────────────────────────────────╮
│ File: /home/user/.config/gocrosshair/config.toml          │
╰───────────────────────────────────────────────────────────╯

Problems found:
  - invalid shape "invalid" (must be one of: cross, dot, circle, cross-dot)

Options:
  [R] Reset to default configuration (backs up current file)
  [Q] Quit application

Choice [R/Q]:
```

Choosing `R` will backup your invalid config and create a fresh default configuration.

## How It Works

1. Connects to the X server using pure Go X11 bindings (no CGO)
2. Queries XRandR for multi-monitor geometry
3. Creates an override-redirect window (bypasses window manager)
4. Uses the XShape extension to:
   - Make only the crosshair visible (transparent background)
   - Make the entire window click-through (input passes to applications below)
5. Draws the crosshair at the selected monitor's center (with optional offset)

## Building for Distribution

```bash
# Optimized build with stripped symbols and version
CGO_ENABLED=0 go build \
  -ldflags="-s -w -X main.version=1.1.0" \
  -trimpath \
  -o gocrosshair .
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [jezek/xgb](https://github.com/jezek/xgb) - Pure Go X11 bindings
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML parser for Go
