package overlay

import (
	"fmt"
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/shape"
	"github.com/jezek/xgb/xproto"

	"gocrosshair/config"
)

// Overlay manages the X11 crosshair overlay window.
type Overlay struct {
	conn      *xgb.Conn
	screen    *xproto.ScreenInfo
	windowID  xproto.Window
	gcID      xproto.Gcontext
	outlineGC xproto.Gcontext
	config    *config.Config
	monitor   Monitor
	centerX   int16
	centerY   int16
}

// NewOverlay creates a new crosshair overlay connected to the X server.
func NewOverlay(cfg *config.Config) (*Overlay, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %w", err)
	}

	// Initialize the shape extension for click-through functionality.
	if err := shape.Init(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize shape extension: %w", err)
	}

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)

	monitors, err := GetMonitors(conn, screen)
	if err != nil {
		log.Printf("Warning: failed to get monitors: %v, using screen dimensions", err)
		monitors = []Monitor{{
			Name:    "default",
			X:       0,
			Y:       0,
			Width:   screen.WidthInPixels,
			Height:  screen.HeightInPixels,
			Primary: true,
		}}
	}

	monitor := SelectMonitor(monitors, cfg.Position.Monitor)

	centerX := monitor.CenterX() + int16(cfg.Position.OffsetX)
	centerY := monitor.CenterY() + int16(cfg.Position.OffsetY)

	return &Overlay{
		conn:    conn,
		screen:  screen,
		config:  cfg,
		monitor: monitor,
		centerX: centerX,
		centerY: centerY,
	}, nil
}

// Close releases X server resources and closes the connection.
func (o *Overlay) Close() {
	if o.conn != nil {
		o.conn.Close()
	}
}

// createWindow creates the overlay window with override-redirect to bypass WM control.
func (o *Overlay) createWindow() error {
	wid, err := xproto.NewWindowId(o.conn)
	if err != nil {
		return fmt.Errorf("failed to create window ID: %w", err)
	}
	o.windowID = wid

	// Create a full-screen window to ensure we can draw anywhere
	screenWidth := o.screen.WidthInPixels
	screenHeight := o.screen.HeightInPixels

	// Window attributes:
	// - OverrideRedirect: bypass window manager (no decorations, absolute positioning)
	// - BackPixel: background color (will be shaped away)
	// - EventMask: we need exposure events for redrawing
	mask := uint32(xproto.CwBackPixel | xproto.CwOverrideRedirect | xproto.CwEventMask)
	values := []uint32{
		0x000000, // BackPixel: black (will be transparent via shape)
		1,        // OverrideRedirect: true
		xproto.EventMaskExposure | xproto.EventMaskStructureNotify,
	}

	err = xproto.CreateWindowChecked(
		o.conn,
		o.screen.RootDepth,
		o.windowID,
		o.screen.Root,
		0, 0,
		screenWidth,
		screenHeight,
		0,
		xproto.WindowClassInputOutput,
		o.screen.RootVisual,
		mask,
		values,
	).Check()

	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	return nil
}

// createGraphicsContext creates graphics contexts for drawing.
func (o *Overlay) createGraphicsContext() error {
	gcid, err := xproto.NewGcontextId(o.conn)
	if err != nil {
		return fmt.Errorf("failed to create GC ID: %w", err)
	}
	o.gcID = gcid

	color := o.config.GetColorUint32()
	mask := uint32(xproto.GcForeground)
	values := []uint32{color}

	if err := xproto.CreateGCChecked(o.conn, o.gcID, xproto.Drawable(o.windowID), mask, values).Check(); err != nil {
		return fmt.Errorf("failed to create GC: %w", err)
	}

	if o.config.Crosshair.OutlineThickness > 0 {
		outlineGC, err := xproto.NewGcontextId(o.conn)
		if err != nil {
			return fmt.Errorf("failed to create outline GC ID: %w", err)
		}
		o.outlineGC = outlineGC

		outlineColor := o.config.GetOutlineColorUint32()
		if err := xproto.CreateGCChecked(o.conn, o.outlineGC, xproto.Drawable(o.windowID), mask, []uint32{outlineColor}).Check(); err != nil {
			return fmt.Errorf("failed to create outline GC: %w", err)
		}
	}

	return nil
}

// applyShape configures the window shape for transparency and click-through.
func (o *Overlay) applyShape() error {
	cfg := o.config.Crosshair

	shapeRects := GenerateShape(
		cfg.Shape,
		o.centerX,
		o.centerY,
		int16(cfg.Size),
		int16(cfg.Thickness),
		int16(cfg.Gap),
	)

	var boundingRects []xproto.Rectangle
	if cfg.OutlineThickness > 0 {
		outlineRects := GenerateOutline(shapeRects, int16(cfg.OutlineThickness))
		boundingRects = append(boundingRects, outlineRects...)
	}
	boundingRects = append(boundingRects, shapeRects...)

	// Set the BOUNDING shape: defines the visible area of the window
	err := shape.RectanglesChecked(
		o.conn,
		shape.SoSet,
		shape.SkBounding,
		xproto.ClipOrderingUnsorted,
		o.windowID,
		0, 0,
		boundingRects,
	).Check()
	if err != nil {
		return fmt.Errorf("failed to set bounding shape: %w", err)
	}

	// Set the INPUT shape: empty = entire window is click-through
	err = shape.RectanglesChecked(
		o.conn,
		shape.SoSet,
		shape.SkInput,
		xproto.ClipOrderingUnsorted,
		o.windowID,
		0, 0,
		[]xproto.Rectangle{},
	).Check()
	if err != nil {
		return fmt.Errorf("failed to set input shape: %w", err)
	}

	return nil
}

// drawCrosshair renders the crosshair onto the window.
func (o *Overlay) drawCrosshair() error {
	cfg := o.config.Crosshair

	shapeRects := GenerateShape(
		cfg.Shape,
		o.centerX,
		o.centerY,
		int16(cfg.Size),
		int16(cfg.Thickness),
		int16(cfg.Gap),
	)

	if cfg.OutlineThickness > 0 && o.outlineGC != 0 {
		outlineRects := GenerateOutline(shapeRects, int16(cfg.OutlineThickness))
		if len(outlineRects) > 0 {
			if err := xproto.PolyFillRectangleChecked(o.conn, xproto.Drawable(o.windowID), o.outlineGC, outlineRects).Check(); err != nil {
				return fmt.Errorf("failed to draw outline: %w", err)
			}
		}
	}

	if len(shapeRects) > 0 {
		if err := xproto.PolyFillRectangleChecked(o.conn, xproto.Drawable(o.windowID), o.gcID, shapeRects).Check(); err != nil {
			return fmt.Errorf("failed to draw crosshair: %w", err)
		}
	}

	return nil
}

// Run initializes and runs the overlay event loop.
func (o *Overlay) Run() error {
	if err := o.createWindow(); err != nil {
		return err
	}

	if err := o.createGraphicsContext(); err != nil {
		return err
	}

	if err := o.applyShape(); err != nil {
		return err
	}

	if err := xproto.MapWindowChecked(o.conn, o.windowID).Check(); err != nil {
		return fmt.Errorf("failed to map window: %w", err)
	}

	if err := o.drawCrosshair(); err != nil {
		return err
	}

	log.Printf("Crosshair overlay running on monitor %q at (%d, %d). Press Ctrl+C to exit.",
		o.monitor.Name, o.centerX, o.centerY)

	for {
		ev, err := o.conn.WaitForEvent()
		if err != nil {
			return fmt.Errorf("X11 connection error: %w", err)
		}

		if ev == nil {
			return nil
		}

		switch ev.(type) {
		case xproto.ExposeEvent:
			if err := o.drawCrosshair(); err != nil {
				log.Printf("Warning: failed to redraw crosshair: %v", err)
			}
		}
	}
}

// ListMonitors connects to X server and prints available monitors.
func ListMonitors() error {
	conn, err := xgb.NewConn()
	if err != nil {
		return fmt.Errorf("failed to connect to X server: %w", err)
	}
	defer conn.Close()

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)

	monitors, err := GetMonitors(conn, screen)
	if err != nil {
		return fmt.Errorf("failed to get monitors: %w", err)
	}

	PrintMonitors(monitors)
	return nil
}
