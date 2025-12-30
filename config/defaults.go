// Package config handles configuration loading, saving, and validation for gocrosshair.
package config

// Default configuration values.
const (
	DefaultShape            = "cross"
	DefaultColor            = "#00FF00"
	DefaultSize             = 10
	DefaultThickness        = 2
	DefaultGap              = 0
	DefaultOutlineThickness = 0
	DefaultOutlineColor     = "#000000"
	DefaultMonitor          = 0
	DefaultOffsetX          = 0
	DefaultOffsetY          = 0
)

// Valid shape options.
var ValidShapes = []string{"cross", "dot", "circle", "cross-dot"}

// Default returns a new Config with default values.
func Default() *Config {
	return &Config{
		Crosshair: CrosshairConfig{
			Shape:            DefaultShape,
			Color:            DefaultColor,
			Size:             DefaultSize,
			Thickness:        DefaultThickness,
			Gap:              DefaultGap,
			OutlineThickness: DefaultOutlineThickness,
			OutlineColor:     DefaultOutlineColor,
		},
		Position: PositionConfig{
			Monitor: DefaultMonitor,
			OffsetX: DefaultOffsetX,
			OffsetY: DefaultOffsetY,
		},
	}
}

// DefaultConfigContent returns the default configuration as a TOML string with comments.
func DefaultConfigContent() string {
	return `# gocrosshair configuration file

[crosshair]
# Shape of the crosshair: "cross", "dot", "circle", "cross-dot"
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
# Monitors are ordered left-to-right by X position
# Use -1 for primary monitor
monitor = 0

# Offset from monitor center (pixels)
# Positive X = right, Positive Y = down
offset_x = 0
offset_y = 0
`
}
