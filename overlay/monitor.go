// Package overlay provides X11 overlay window functionality for the crosshair.
package overlay

import (
	"fmt"
	"sort"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
)

// Monitor represents a display output with its geometry.
type Monitor struct {
	Name      string
	X         int16
	Y         int16
	Width     uint16
	Height    uint16
	Primary   bool
	Connected bool
}

// CenterX returns the X coordinate of the monitor's center.
func (m Monitor) CenterX() int16 {
	return m.X + int16(m.Width/2)
}

// CenterY returns the Y coordinate of the monitor's center.
func (m Monitor) CenterY() int16 {
	return m.Y + int16(m.Height/2)
}

// String returns a human-readable representation of the monitor.
func (m Monitor) String() string {
	primary := ""
	if m.Primary {
		primary = " (primary)"
	}
	return fmt.Sprintf("%s: %dx%d+%d+%d%s",
		m.Name, m.Width, m.Height, m.X, m.Y, primary)
}

// GetMonitors queries XRandR for available monitors.
func GetMonitors(conn *xgb.Conn, screen *xproto.ScreenInfo) ([]Monitor, error) {
	if err := randr.Init(conn); err != nil {
		return nil, fmt.Errorf("failed to initialize RandR: %w", err)
	}

	resources, err := randr.GetScreenResources(conn, screen.Root).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get screen resources: %w", err)
	}

	primaryReply, err := randr.GetOutputPrimary(conn, screen.Root).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get primary output: %w", err)
	}
	primaryOutput := primaryReply.Output

	var monitors []Monitor

	for _, output := range resources.Outputs {
		outputInfo, err := randr.GetOutputInfo(conn, output, resources.ConfigTimestamp).Reply()
		if err != nil {
			continue
		}

		if outputInfo.Connection != randr.ConnectionConnected {
			continue
		}

		if outputInfo.Crtc == 0 {
			continue
		}

		crtcInfo, err := randr.GetCrtcInfo(conn, outputInfo.Crtc, resources.ConfigTimestamp).Reply()
		if err != nil {
			continue
		}

		if crtcInfo.Width == 0 || crtcInfo.Height == 0 {
			continue
		}

		mon := Monitor{
			Name:      string(outputInfo.Name),
			X:         crtcInfo.X,
			Y:         crtcInfo.Y,
			Width:     crtcInfo.Width,
			Height:    crtcInfo.Height,
			Primary:   output == primaryOutput,
			Connected: true,
		}

		monitors = append(monitors, mon)
	}

	sort.Slice(monitors, func(i, j int) bool {
		if monitors[i].X != monitors[j].X {
			return monitors[i].X < monitors[j].X
		}
		return monitors[i].Y < monitors[j].Y
	})

	if len(monitors) == 0 {
		monitors = append(monitors, Monitor{
			Name:      "default",
			X:         0,
			Y:         0,
			Width:     screen.WidthInPixels,
			Height:    screen.HeightInPixels,
			Primary:   true,
			Connected: true,
		})
	}

	return monitors, nil
}

// SelectMonitor selects a monitor by index.
// Index -1 selects the primary monitor.
// If index is out of range, falls back to the first monitor.
func SelectMonitor(monitors []Monitor, index int) Monitor {
	if len(monitors) == 0 {
		return Monitor{Width: 1920, Height: 1080}
	}

	if index == -1 {
		for _, m := range monitors {
			if m.Primary {
				return m
			}
		}
		return monitors[0]
	}

	if index >= 0 && index < len(monitors) {
		return monitors[index]
	}

	return monitors[0]
}

// PrintMonitors outputs a formatted list of monitors.
func PrintMonitors(monitors []Monitor) {
	fmt.Println("Available monitors:")
	fmt.Println()

	for i, m := range monitors {
		primary := ""
		if m.Primary {
			primary = " ← primary"
		}
		fmt.Printf("  [%d] %s: %dx%d at position (%d, %d)%s\n",
			i, m.Name, m.Width, m.Height, m.X, m.Y, primary)
	}

	fmt.Println()
	fmt.Println("Use 'monitor = N' in config to select a monitor by index.")
	fmt.Println("Use 'monitor = -1' to automatically select the primary monitor.")
}
