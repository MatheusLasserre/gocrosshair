# Configuration System Implementation Plan

## Overview

Add a TOML-based configuration system that allows users to customize the crosshair appearance, positioning, and target monitor. The application will load an existing config file or generate a default one if none exists.

---

## Phase 1: Configuration File Structure

### File Location Strategy

Follow XDG Base Directory specification:
1. Check `$XDG_CONFIG_HOME/gocrosshair/config.toml`
2. Fallback to `~/.config/gocrosshair/config.toml`
3. Allow override via `-config <path>` CLI flag

### Configuration Schema (TOML)

```toml
# gocrosshair configuration file

[crosshair]
# Shape of the crosshair: "cross", "dot", "circle", "cross-dot"
shape = "cross"

# Color in hex format (0xRRGGBB or #RRGGBB)
color = "#00FF00"

# Size of the crosshair arms (pixels from center)
size = 10

# Thickness of lines (pixels)
thickness = 2

# Gap in center (pixels) - for hollow cross shapes
gap = 0

# Outline settings (0 to disable)
outline_thickness = 0
outline_color = "#000000"

[position]
# Monitor index (0 = primary, 1 = secondary, etc.)
# Use -1 for automatic (primary monitor)
monitor = 0

# Offset from monitor center (pixels)
# Positive X = right, Positive Y = down
offset_x = 0
offset_y = 0
```

---

## Phase 2: Dependencies

### Required Package

Add TOML parser (zero CGO, pure Go):
```bash
go get github.com/BurntSushi/toml
```

### Required X Extension

Add XRandR for multi-monitor support:
```bash
go get github.com/jezek/xgb/randr
```

---

## Phase 3: Project Structure

Refactor into a clean package structure:

```
gocrosshair/
├── main.go              # Entry point, CLI parsing
├── config/
│   ├── config.go        # Config struct, Load/Save logic
│   └── defaults.go      # Default configuration values
├── overlay/
│   ├── overlay.go       # X11 overlay window management
│   ├── shapes.go        # Crosshair shape generators
│   └── monitor.go       # Multi-monitor detection (XRandR)
├── go.mod
├── go.sum
├── PKGBUILD
├── README.md
└── LICENSE
```

---

## Phase 4: Implementation Details

### 4.1 Config Package (`config/config.go`)

```go
type Config struct {
    Crosshair CrosshairConfig `toml:"crosshair"`
    Position  PositionConfig  `toml:"position"`
}

type CrosshairConfig struct {
    Shape            string `toml:"shape"`
    Color            string `toml:"color"`
    Size             int    `toml:"size"`
    Thickness        int    `toml:"thickness"`
    Gap              int    `toml:"gap"`
    OutlineThickness int    `toml:"outline_thickness"`
    OutlineColor     string `toml:"outline_color"`
}

type PositionConfig struct {
    Monitor int `toml:"monitor"`
    OffsetX int `toml:"offset_x"`
    OffsetY int `toml:"offset_y"`
}
```

**Functions to implement:**
- `GetConfigPath() string` - Resolve config file path (XDG spec)
- `Load(path string) (*Config, error)` - Load and validate config
- `Save(path string, cfg *Config) error` - Write config to file
- `LoadOrCreate(path string) (*Config, error)` - Load existing or create default
- `Default() *Config` - Return default configuration
- `(c *Config) Validate() error` - Validate configuration values
- `ParseColor(s string) (uint32, error)` - Parse hex color string to uint32

### 4.2 Monitor Detection (`overlay/monitor.go`)

**Functions to implement:**
- `type Monitor struct` - Monitor geometry (X, Y, Width, Height, Name, Primary)
- `GetMonitors(conn *xgb.Conn) ([]Monitor, error)` - Query XRandR for monitors
- `GetMonitorCenter(m Monitor) (x, y int16)` - Calculate monitor center point
- `SelectMonitor(monitors []Monitor, index int) Monitor` - Select by index with fallback

**XRandR Logic:**
1. Initialize RandR extension
2. Query screen resources
3. Iterate CRTCs to get active monitor geometries
4. Sort by X position (left-to-right ordering)
5. Mark primary monitor

### 4.3 Shape Generation (`overlay/shapes.go`)

**Supported Shapes:**

1. **Cross** - Traditional + shape
2. **Dot** - Single filled square at center
3. **Circle** - Filled circle at center (approximated with rectangles)
4. **CrossDot** - Cross with center dot

**Circle Approximation:**
Since X11 shapes only support rectangles, we approximate a filled circle using horizontal rectangle slices (scanlines). For a circle of radius R, we calculate the width at each Y position using the circle equation: `width = 2 * sqrt(R² - y²)`

**Interface:**
```go
type ShapeGenerator interface {
    Generate(centerX, centerY int16, cfg *config.CrosshairConfig) []xproto.Rectangle
}
```

**Functions:**
- `GenerateCross(centerX, centerY, size, thickness, gap int16) []xproto.Rectangle`
- `GenerateDot(centerX, centerY, size int16) []xproto.Rectangle` - Square dot
- `GenerateCircle(centerX, centerY, radius int16) []xproto.Rectangle` - Filled circle via scanlines
- `GenerateCrossDot(centerX, centerY, size, thickness, dotSize int16) []xproto.Rectangle`

### 4.4 Overlay Refactor (`overlay/overlay.go`)

Update `Overlay` struct to accept configuration:

```go
type Overlay struct {
    conn     *xgb.Conn
    screen   *xproto.ScreenInfo
    windowID xproto.Window
    gcID     xproto.Gcontext
    config   *config.Config
    monitor  Monitor
    centerX  int16
    centerY  int16
}

func NewOverlay(cfg *config.Config) (*Overlay, error)
```

**Changes:**
- Calculate center based on selected monitor + offset
- Use config values for shape, color, size, thickness
- Position window to cover selected monitor only (optimization)

### 4.5 CLI Parsing (`main.go`)

**Flags:**
- `-config <path>` - Custom config file path (overrides default location)
- `-list-monitors` - List available monitors and exit
- `-version` - Print version and exit

### 4.6 Startup Flow & Config Validation

**Initialization Sequence:**

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Start                         │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌───────────────────────┐
                    │  Resolve config path  │
                    │  (flag or XDG default)│
                    └───────────────────────┘
                                 │
                                 ▼
                    ┌───────────────────────┐
                    │  Config file exists?  │
                    └───────────────────────┘
                           │           │
                          No          Yes
                           │           │
                           ▼           ▼
              ┌─────────────────┐  ┌─────────────────┐
              │ Create default  │  │  Parse config   │
              │ config + notify │  │     file        │
              └─────────────────┘  └─────────────────┘
                           │           │
                           │           ▼
                           │  ┌─────────────────────┐
                           │  │  Validation passed? │
                           │  └─────────────────────┘
                           │        │           │
                           │       Yes          No
                           │        │           │
                           │        │           ▼
                           │        │  ┌─────────────────────────┐
                           │        │  │  Display error details  │
                           │        │  │  Prompt user:           │
                           │        │  │  [R]eset to defaults    │
                           │        │  │  [Q]uit                 │
                           │        │  └─────────────────────────┘
                           │        │           │
                           │        │      R    │    Q
                           │        │      │    │    │
                           │        │      ▼    │    ▼
                           │        │  ┌──────┐ │  ┌──────┐
                           │        │  │Reset │ │  │Exit  │
                           │        │  │config│ │  │(1)   │
                           │        │  └──────┘ │  └──────┘
                           │        │      │    │
                           ▼        ▼      ▼    │
              ┌─────────────────────────────────┐
              │        Run overlay              │
              └─────────────────────────────────┘
```

**Validation Checks:**
- TOML syntax is valid
- Shape is one of: "cross", "dot", "circle", "cross-dot"
- Color is valid hex format
- Size, thickness are positive integers
- Monitor index is >= -1
- Numeric values are within reasonable bounds

**Error Handling Function:**
```go
func handleInvalidConfig(path string, err error) (*Config, error) {
    fmt.Fprintf(os.Stderr, "Configuration error in %s:\n", path)
    fmt.Fprintf(os.Stderr, "  %v\n\n", err)
    fmt.Fprintf(os.Stderr, "Options:\n")
    fmt.Fprintf(os.Stderr, "  [R] Reset to default configuration\n")
    fmt.Fprintf(os.Stderr, "  [Q] Quit application\n")
    fmt.Fprintf(os.Stderr, "\nChoice [R/Q]: ")
    
    // Read single character input
    // If 'R' or 'r': backup old config, write defaults, return default config
    // If 'Q' or 'q' or other: return error to exit
}
```

**Config Backup on Reset:**
When user chooses to reset, rename the invalid config:
`config.toml` → `config.toml.bak.{timestamp}`

---

## Phase 5: Implementation Order

### Step 1: Add Dependencies
```bash
go get github.com/BurntSushi/toml
go get github.com/jezek/xgb/randr
```

### Step 2: Create `config/` Package
1. Create `config/defaults.go` with default values
2. Create `config/config.go` with struct and Load/Save logic
3. Add color parsing utility
4. Add comprehensive validation logic
5. Add interactive error handling with reset option

### Step 3: Create Monitor Detection
1. Create `overlay/monitor.go`
2. Implement XRandR queries
3. Add monitor selection logic

### Step 4: Create Shape Generators
1. Create `overlay/shapes.go`
2. Implement cross shape with gap support
3. Implement square dot shape
4. Implement filled circle using scanline approximation
5. Implement cross-dot combination
6. Add outline support (draw outline rects first, then inner)

### Step 5: Refactor Overlay
1. Move to `overlay/overlay.go`
2. Update to use Config struct
3. Update to use selected monitor geometry
4. Update drawing to use shape generators

### Step 6: Update Main
1. Add flag parsing
2. Add config loading with auto-creation logic
3. Add validation with interactive error recovery
4. Add `-list-monitors` command

### Step 7: Testing & Documentation
1. Test each shape type
2. Test multi-monitor scenarios
3. Test config file generation
4. Update README.md

---

## Phase 6: Example Configurations

### Minimal CS-style Crosshair
```toml
[crosshair]
shape = "cross"
color = "#00FF00"
size = 5
thickness = 2
gap = 3

[position]
monitor = 0
offset_x = 0
offset_y = 0
```

### Center Dot Only
```toml
[crosshair]
shape = "dot"
color = "#FF0000"
size = 4
thickness = 0

[position]
monitor = 0
offset_x = 0
offset_y = 0
```

### Cross with Outline
```toml
[crosshair]
shape = "cross"
color = "#FFFFFF"
size = 12
thickness = 2
gap = 4
outline_thickness = 1
outline_color = "#000000"

[position]
monitor = 1
offset_x = 0
offset_y = 0
```

---

## Phase 7: Build & Distribution

### Updated Build Command
```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o gocrosshair ./...
```

### PKGBUILD Updates
No changes needed - still pure Go with no CGO dependencies.

---

## Technical Notes

### XRandR Considerations
- XRandR is available on XWayland
- CRTCs represent active display outputs
- Multiple monitors may share the same CRTC (mirrored)
- Primary output is flagged in RandR 1.3+

### Color Parsing
Support formats:
- `#RRGGBB` (web format)
- `0xRRGGBB` (hex literal)
- `RRGGBB` (raw hex)

### Shape Rectangles for Circle
X11 shapes only support rectangles. We approximate a filled circle using horizontal scanlines:
- For radius R, iterate Y from -R to +R
- At each Y, calculate width: `w = 2 * sqrt(R² - Y²)`
- Create a rectangle of height 1 at that Y position with calculated width
- This produces a smooth filled circle appearance

Example for radius 5:
```
     ████       (y=-4, w=6)
   ████████     (y=-3, w=8)
  ██████████    (y=-2, w=9)
  ██████████    (y=-1, w=9)
 ████████████   (y=0,  w=10)
  ██████████    (y=1,  w=9)
  ██████████    (y=2,  w=9)
   ████████     (y=3,  w=8)
     ████       (y=4,  w=6)
```

### Config Hot-Reload (Future Enhancement)
Could add SIGHUP handler to reload config without restart.

---

## Success Criteria

1. ✅ Binary accepts `-config <path>` flag for custom config location
2. ✅ Default config auto-generated at `~/.config/gocrosshair/config.toml` if missing
3. ✅ Invalid config prompts user: reset to defaults or quit
4. ✅ All shape types render correctly (cross, dot, circle, cross-dot)
5. ✅ Filled circle renders smoothly via scanline approximation
6. ✅ Monitor selection works on multi-monitor setups
7. ✅ Offset positioning works as expected
8. ✅ Config validation with clear error messages
9. ✅ `-list-monitors` shows available displays
10. ✅ Maintains <3MB binary size
11. ✅ Zero CGO dependencies
