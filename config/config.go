// Package config handles configuration loading, saving, and validation for gocrosshair.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the complete application configuration.
type Config struct {
	Crosshair CrosshairConfig `toml:"crosshair"`
	Position  PositionConfig  `toml:"position"`
}

// CrosshairConfig contains crosshair appearance settings.
type CrosshairConfig struct {
	Shape            string `toml:"shape"`
	Color            string `toml:"color"`
	Size             int    `toml:"size"`
	Thickness        int    `toml:"thickness"`
	Gap              int    `toml:"gap"`
	OutlineThickness int    `toml:"outline_thickness"`
	OutlineColor     string `toml:"outline_color"`
}

// PositionConfig contains crosshair positioning settings.
type PositionConfig struct {
	Monitor int `toml:"monitor"`
	OffsetX int `toml:"offset_x"`
	OffsetY int `toml:"offset_y"`
}

// GetConfigDir returns the configuration directory path following XDG spec.
func GetConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "gocrosshair")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home is unavailable
		return ".gocrosshair"
	}

	return filepath.Join(home, ".config", "gocrosshair")
}

// GetConfigPath returns the full path to the configuration file.
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.toml")
}

// Load reads and parses a configuration file from the given path.
func Load(path string) (*Config, error) {
	cfg := &Config{}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to the given path.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveDefault writes the default configuration with comments to the given path.
func SaveDefault(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(DefaultConfigContent()), 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// LoadOrCreate loads an existing config or creates a default one.
// Returns the config and a boolean indicating if a new config was created.
func LoadOrCreate(path string) (*Config, bool, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		if err := SaveDefault(path); err != nil {
			return nil, false, err
		}
		fmt.Printf("Created default configuration at: %s\n", path)
		return Default(), true, nil
	}

	// Load existing config
	cfg, err := Load(path)
	if err != nil {
		return nil, false, err
	}

	return cfg, false, nil
}

// Validate checks if the configuration values are valid.
func (c *Config) Validate() error {
	var errs []string

	if !slices.Contains(ValidShapes, c.Crosshair.Shape) {
		errs = append(errs, fmt.Sprintf("invalid shape %q (must be one of: %s)",
			c.Crosshair.Shape, strings.Join(ValidShapes, ", ")))
	}

	if _, err := ParseColor(c.Crosshair.Color); err != nil {
		errs = append(errs, fmt.Sprintf("invalid color %q: %v", c.Crosshair.Color, err))
	}

	if _, err := ParseColor(c.Crosshair.OutlineColor); err != nil {
		errs = append(errs, fmt.Sprintf("invalid outline_color %q: %v", c.Crosshair.OutlineColor, err))
	}

	if c.Crosshair.Size < 1 || c.Crosshair.Size > 500 {
		errs = append(errs, fmt.Sprintf("size must be between 1 and 500 (got %d)", c.Crosshair.Size))
	}

	if c.Crosshair.Thickness < 1 || c.Crosshair.Thickness > 100 {
		errs = append(errs, fmt.Sprintf("thickness must be between 1 and 100 (got %d)", c.Crosshair.Thickness))
	}

	if c.Crosshair.Gap < 0 || c.Crosshair.Gap > 100 {
		errs = append(errs, fmt.Sprintf("gap must be between 0 and 100 (got %d)", c.Crosshair.Gap))
	}

	if c.Crosshair.OutlineThickness < 0 || c.Crosshair.OutlineThickness > 50 {
		errs = append(errs, fmt.Sprintf("outline_thickness must be between 0 and 50 (got %d)", c.Crosshair.OutlineThickness))
	}

	if c.Position.Monitor < -1 || c.Position.Monitor > 100 {
		errs = append(errs, fmt.Sprintf("monitor must be between -1 and 100 (got %d)", c.Position.Monitor))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n  - "))
	}

	return nil
}

// ParseColor parses a hex color string and returns the uint32 value.
// Supports formats: #RRGGBB, 0xRRGGBB, RRGGBB
func ParseColor(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	if len(s) != 6 {
		return 0, fmt.Errorf("color must be 6 hex digits (got %q)", s)
	}

	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid hex color: %w", err)
	}

	return uint32(val), nil
}

// HandleInvalidConfig prompts the user to reset or quit when config is invalid.
// Returns the default config if user chooses to reset, or an error to quit.
func HandleInvalidConfig(path string, validationErr error) (*Config, error) {
	fmt.Fprintf(os.Stderr, "\n╭─ Configuration Error ─────────────────────────────────────╮\n")
	fmt.Fprintf(os.Stderr, "│ File: %-53s │\n", path)
	fmt.Fprintf(os.Stderr, "╰───────────────────────────────────────────────────────────╯\n\n")
	fmt.Fprintf(os.Stderr, "Problems found:\n  - %v\n\n", validationErr)
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  [R] Reset to default configuration (backs up current file)\n")
	fmt.Fprintf(os.Stderr, "  [Q] Quit application\n\n")
	fmt.Fprintf(os.Stderr, "Choice [R/Q]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "r" {
		backupPath := fmt.Sprintf("%s.bak.%d", path, time.Now().Unix())
		if err := os.Rename(path, backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not backup config: %v\n", err)
		} else {
			fmt.Printf("Backed up invalid config to: %s\n", backupPath)
		}

		if err := SaveDefault(path); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}

		fmt.Printf("Created fresh configuration at: %s\n", path)
		return Default(), nil
	}

	return nil, errors.New("user chose to quit")
}

// GetColorUint32 returns the crosshair color as uint32.
func (c *Config) GetColorUint32() uint32 {
	color, _ := ParseColor(c.Crosshair.Color)
	return color
}

// GetOutlineColorUint32 returns the outline color as uint32.
func (c *Config) GetOutlineColorUint32() uint32 {
	color, _ := ParseColor(c.Crosshair.OutlineColor)
	return color
}
