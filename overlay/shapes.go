package overlay

import (
	"github.com/jezek/xgb/xproto"
)

// GenerateCross creates rectangles for a cross/plus shape.
// gap specifies the size of the center gap (0 for solid cross).
func GenerateCross(centerX, centerY, size, thickness, gap int16) []xproto.Rectangle {
	halfThickness := thickness / 2
	rects := make([]xproto.Rectangle, 0, 4)

	if gap <= 0 {
		// Solid cross - two rectangles
		// Horizontal line
		rects = append(rects, xproto.Rectangle{
			X:      centerX - size,
			Y:      centerY - halfThickness,
			Width:  uint16(size * 2),
			Height: uint16(thickness),
		})
		// Vertical line
		rects = append(rects, xproto.Rectangle{
			X:      centerX - halfThickness,
			Y:      centerY - size,
			Width:  uint16(thickness),
			Height: uint16(size * 2),
		})
	} else {
		// Cross with gap - four rectangles
		halfGap := gap / 2

		// Left arm
		rects = append(rects, xproto.Rectangle{
			X:      centerX - size,
			Y:      centerY - halfThickness,
			Width:  uint16(size - halfGap),
			Height: uint16(thickness),
		})
		// Right arm
		rects = append(rects, xproto.Rectangle{
			X:      centerX + halfGap,
			Y:      centerY - halfThickness,
			Width:  uint16(size - halfGap),
			Height: uint16(thickness),
		})
		// Top arm
		rects = append(rects, xproto.Rectangle{
			X:      centerX - halfThickness,
			Y:      centerY - size,
			Width:  uint16(thickness),
			Height: uint16(size - halfGap),
		})
		// Bottom arm
		rects = append(rects, xproto.Rectangle{
			X:      centerX - halfThickness,
			Y:      centerY + halfGap,
			Width:  uint16(thickness),
			Height: uint16(size - halfGap),
		})
	}

	return rects
}

// makeCenteredLine creates a horizontal rectangle centered at (cx, cy).
// The total width is (offset * 2) + 1, ensuring the line always has
// a specific center pixel, preventing "wobbly" circles.
func makeCenteredLine(cx, cy, offset int16) xproto.Rectangle {
	return xproto.Rectangle{
		X:      cx - offset,
		Y:      cy,
		Width:  uint16(offset*2 + 1), // Always odd width = perfect center
		Height: 1,
	}
}

// GenerateDot creates a circular dot at the center using the Midpoint Circle Algorithm.
// This ensures pixel-perfect symmetry with a distinct center pixel.
func GenerateDot(centerX, centerY, size int16) []xproto.Rectangle {
	if size <= 0 {
		return nil
	}

	// Use radius as half the size for a circular dot
	radius := size / 2
	if radius < 1 {
		// For very small sizes, just return a single pixel
		return []xproto.Rectangle{{
			X:      centerX,
			Y:      centerY,
			Width:  1,
			Height: 1,
		}}
	}

	return generateFilledCircle(centerX, centerY, radius)
}

// GenerateCircle creates a filled circle using the Midpoint Circle Algorithm.
// It ensures perfect symmetry by forcing all scanlines to have an odd width (2*x + 1),
// creating a distinct "center pixel" which is crucial for crosshairs.
func GenerateCircle(centerX, centerY, radius int16) []xproto.Rectangle {
	if radius <= 0 {
		return nil
	}

	return generateFilledCircle(centerX, centerY, radius)
}

// generateFilledCircle implements the Midpoint Circle Algorithm (Bresenham).
// This uses purely integer math to select the most visually pleasing pixels
// and ensures perfect symmetry.
func generateFilledCircle(centerX, centerY, radius int16) []xproto.Rectangle {
	var rects []xproto.Rectangle

	// x and y represent the offset from the center
	x := int16(0)
	y := radius
	d := 3 - 2*radius // Decision parameter

	for y >= x {
		// Because a circle is symmetrical in octants, we can draw 4 lines per loop
		// to fill the circle.

		// 1. Horizontal bands at the top and bottom of the circle
		// These are wide lines. We use 'y' as the horizontal offset.
		if x > 0 {
			// Upper band
			rects = append(rects, makeCenteredLine(centerX, centerY-x, y))
			// Lower band (mirror)
			rects = append(rects, makeCenteredLine(centerX, centerY+x, y))
		} else {
			// Center line (only add once when x == 0)
			rects = append(rects, makeCenteredLine(centerX, centerY, y))
		}

		// 2. Horizontal bands in the middle section
		// These are narrower lines. We use 'x' as the horizontal offset.
		// Check x != y to prevent overlapping lines at the 45-degree mark
		if x != y {
			// Upper middle
			rects = append(rects, makeCenteredLine(centerX, centerY-y, x))
			// Lower middle
			rects = append(rects, makeCenteredLine(centerX, centerY+y, x))
		}

		// Update decision parameter (Bresenham's logic)
		if d > 0 {
			y--
			d = d + 4*(x-y) + 10
		} else {
			d = d + 4*x + 6
		}
		x++
	}

	return rects
}

// GenerateCrossDot creates a cross with a center dot.
func GenerateCrossDot(centerX, centerY, size, thickness, gap, dotSize int16) []xproto.Rectangle {
	effectiveGap := gap
	if effectiveGap < dotSize {
		effectiveGap = dotSize
	}

	rects := GenerateCross(centerX, centerY, size, thickness, effectiveGap)

	dotRects := GenerateDot(centerX, centerY, dotSize)
	rects = append(rects, dotRects...)

	return rects
}

// GenerateOutline creates outline rectangles around the given shape.
// It generates a larger version of the shape that will be drawn behind the main shape.
func GenerateOutline(rects []xproto.Rectangle, outlineThickness int16) []xproto.Rectangle {
	if outlineThickness <= 0 {
		return nil
	}

	outlineRects := make([]xproto.Rectangle, len(rects))
	for i, r := range rects {
		outlineRects[i] = xproto.Rectangle{
			X:      r.X - outlineThickness,
			Y:      r.Y - outlineThickness,
			Width:  r.Width + uint16(outlineThickness*2),
			Height: r.Height + uint16(outlineThickness*2),
		}
	}

	return outlineRects
}

// GenerateShape creates rectangles for the specified shape type.
func GenerateShape(shape string, centerX, centerY, size, thickness, gap int16) []xproto.Rectangle {
	switch shape {
	case "cross":
		return GenerateCross(centerX, centerY, size, thickness, gap)
	case "dot":
		return GenerateDot(centerX, centerY, size)
	case "circle":
		return GenerateCircle(centerX, centerY, size)
	case "cross-dot":
		dotSize := max(size/3, 2)
		return GenerateCrossDot(centerX, centerY, size, thickness, gap, dotSize)
	default:
		return GenerateCross(centerX, centerY, size, thickness, gap)
	}
}
