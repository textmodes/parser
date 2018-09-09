package bgi

import (
	"image"
	"image/color"
)

type WriteMode byte

// Write modes
const (
	WriteModeSet WriteMode = iota
	WriteModeXOR
	WriteModeOR
	WriteModeAND
	WriteModeNOT
)

// Canvas is a Borland Graphics Interface image canvas.
type Canvas struct {
	image.Paletted

	position  image.Point
	writeMode WriteMode
	viewport  image.Rectangle

	line struct {
		path         Path
		patternIndex uint8
		patterns     [5]uint16
	}
	fill struct {
		path         Path
		color        uint8
		patternIndex uint8
		patterns     [13][]byte
	}

	fg, bg  uint8
	hasPath bool
}

func NewCanvas() *Canvas {
	canvas := &Canvas{
		Paletted: image.Paletted{
			Pix:     make([]uint8, 640*350),
			Stride:  1,
			Rect:    image.Rect(0, 0, 640, 350),
			Palette: copyPalette(Palette),
		},
		viewport: image.Rect(0, 0, 640, 350),
		fg:       7,
		bg:       0,
	}
	copy(canvas.line.patterns[:], LinePatterns[:])
	copy(canvas.fill.patterns[:], FillPatterns[:])
	canvas.Clear()
	return canvas
}

// Color sets the foreground color by index
func (canvas *Canvas) Color(index uint8) {
	canvas.fg = index & 0x0f
}

// BackgroundColor sets the background color by index
func (canvas *Canvas) BackgroundColor(index uint8) {
	canvas.bg = index & 0x0f
}

// Clear the canvas with the current background color
func (canvas *Canvas) Clear() {
	for i := range canvas.Pix {
		canvas.Pix[i] = canvas.bg
	}
}

// set a pixel color using the current write mode
func (canvas *Canvas) setColorIndex(x, y int, c uint8) {
	if !image.Pt(x, y).In(canvas.viewport) {
		return
	}

	c &= 0x0f
	offset := canvas.PixOffset(x, y)
	switch canvas.writeMode {
	case WriteModeSet:
		canvas.Pix[offset] = c
	case WriteModeXOR:
		canvas.Pix[offset] ^= c
	case WriteModeOR:
		canvas.Pix[offset] |= c
	case WriteModeAND:
		canvas.Pix[offset] &= c
	case WriteModeNOT:
		canvas.Pix[offset] &= (^c & 0x0f)
	}
}

// set a pixel color respecting the current line pattern
func (canvas *Canvas) setColorLine(x, y int) {
	if !image.Pt(x, y).In(canvas.viewport) {
		return
	}

	if canvas.line.patterns[canvas.line.patternIndex]>>uint(x%8)&1 == 1 {
		canvas.setColorIndex(x, y, canvas.fg)
	} else {
		canvas.setColorIndex(x, y, canvas.bg)
	}
}

// set a pixel color respecting the current fill pattern
func (canvas *Canvas) setColorFill(x, y int) {
	if !image.Pt(x, y).In(canvas.viewport) {
		return
	}

	if canvas.fill.patterns[canvas.fill.patternIndex][y%8]>>uint(x%8)&1 == 1 {
		canvas.setColorIndex(x, y, canvas.fill.color)
	} else {
		canvas.setColorIndex(x, y, canvas.bg)
	}
}

// BGI compatible functions

func (canvas *Canvas) ClearPath() {
	canvas.line.path = canvas.line.path[:0]
	canvas.fill.path = canvas.fill.path[:0]
	canvas.hasPath = false
}

func (canvas *Canvas) MoveTo(x, y int) {
	pt := image.Pt(x, y)
	canvas.line.path = NewPath(pt)
	canvas.fill.path = NewPath(pt)
	canvas.hasPath = true
}

func (canvas *Canvas) LineTo(x, y int) {
	if !canvas.hasPath {
		canvas.MoveTo(x, y)
		return
	}
	pt := image.Pt(x, y)
	canvas.line.path.Add(pt)
	canvas.fill.path.Add(pt)
}

func (canvas *Canvas) stroke(a, b image.Point) {
	var (
		dx, dy, e, slope int
		x1               = a.X
		x2               = b.X
		y1               = a.Y
		y2               = b.Y
	)

	if x1 > x2 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}

	dx, dy = x2-x1, y2-y1

	// we do x-axis scans; dy cannot be negative
	if dy < 0 {
		dy = -dy
	}

	switch {
	case x1 == x2 && y1 == y2:
		// just one pixel
		canvas.setColorLine(x1, y1)

	case y1 == y2:
		// horizontal line
		for ; dx != 0; dx-- {
			canvas.setColorLine(x1, y1)
			x1++
		}
		canvas.setColorLine(x1, y1)

	case x1 == x2:
		// vertical line
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for ; dy != 0; dy-- {
			canvas.setColorLine(x1, y1)
			y1++
		}
		canvas.setColorLine(x1, y1)

	case dx == dy:
		// diagonal line
		if y1 < y2 {
			for ; dx != 0; dx-- {
				canvas.setColorLine(x1, y1)
				x1++
				y1++
			}
		} else {
			for ; dx != 0; dx-- {
				canvas.setColorLine(x1, y1)
				x1++
				y1--
			}
		}
		canvas.setColorLine(x1, y1)

	case dx > dy:
		// wide line
		if y1 < y2 {
			// BresenhamDxXRYD(img, x1, y1, x2, y2, col)
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				canvas.setColorLine(x1, y1)
				x1++
				e -= dy
				if e < 0 {
					y1++
					e += slope
				}
			}
		} else {
			// BresenhamDxXRYU(img, x1, y1, x2, y2, col)
			dy, e, slope = 2*dy, dx, 2*dx
			for ; dx != 0; dx-- {
				canvas.setColorLine(x1, y1)
				x1++
				e -= dy
				if e < 0 {
					y1--
					e += slope
				}
			}
		}
		canvas.setColorLine(x1, y1)

	default:
		// tall line
		if y1 < y2 {
			// BresenhamDyXRYD(img, x1, y1, x2, y2, col)
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				canvas.setColorLine(x1, y1)
				y1++
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		} else {
			// BresenhamDyXRYU(img, x1, y1, x2, y2, col)
			dx, e, slope = 2*dx, dy, 2*dy
			for ; dy != 0; dy-- {
				canvas.setColorLine(x1, y1)
				y1--
				e -= dx
				if e < 0 {
					x1++
					e += slope
				}
			}
		}
		canvas.setColorLine(x1, y1)
	}
}

func (canvas *Canvas) strokeRaster() {
	// TODO: line thickness is not implemented
	for i, b := range canvas.line.path[1:] {
		canvas.stroke(canvas.line.path[i], b)
	}
}

func (canvas *Canvas) StrokePreserve() {
	canvas.strokeRaster()
}

func (canvas *Canvas) Stroke() {
	canvas.StrokePreserve()
	canvas.ClearPath()
}

func (canvas *Canvas) fillRaster() {
	var (
		b = canvas.fill.path.Bounds()
		p image.Point
	)
	for p.Y = b.Min.Y; p.Y < b.Max.Y; p.Y++ {
		for p.X = b.Min.X; p.X < b.Max.X; p.X++ {
			if canvas.fill.path.Contains(p) {
				canvas.setColorFill(p.X, p.Y)
			}
		}
	}
}

func (canvas *Canvas) FillPreserve() {
	canvas.fillRaster()
}

func (canvas *Canvas) Fill() {
	canvas.FillPreserve()
	canvas.ClearPath()
}

// EGAColor returns the indexed EGA color as RGBA color
func EGAColor(index uint8) color.Color {
	return color.RGBA{
		R: (((index & 0x20) >> 5) + ((index & 0x04) >> 1)) * 0x55,
		G: (((index & 0x10) >> 4) + ((index & 0x02) >> 0)) * 0x55,
		B: (((index & 0x08) >> 3) + ((index & 0x01) << 1)) * 0x55,
		A: 0xff,
	}
}

var (
	// Palette is the default BGI palette.
	Palette = color.Palette{
		EGAColor(0),
		EGAColor(1),
		EGAColor(2),
		EGAColor(3),
		EGAColor(4),
		EGAColor(5),
		EGAColor(20),
		EGAColor(7),
		EGAColor(56),
		EGAColor(57),
		EGAColor(58),
		EGAColor(59),
		EGAColor(60),
		EGAColor(61),
		EGAColor(62),
		EGAColor(63),
	}

	// LinePatterns are the default line patterns.
	LinePatterns = [5]uint16{
		0xffff, // solid  0b1111111111111111
		0xcccc, // dotted 0b1100110011001100
		0xf1f8, // center 0b1111000111111000
		0xf8f8, // dashed 0b1111100011111000
		0xffff, // user
	}

	// FillPatterns are the default fill patterns.
	FillPatterns = [13][]byte{
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // EMPTY_FILL
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // SOLID_FILL
		[]byte{0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00}, // LINE_FILL
		[]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}, // LTSLASH_FILL
		[]byte{0x03, 0x06, 0x0c, 0x18, 0x30, 0x60, 0xc0, 0x81}, // SLASH_FILL
		[]byte{0xc0, 0x60, 0x30, 0x18, 0x0c, 0x06, 0x03, 0x81}, // BKSLASH_FILL
		[]byte{0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01}, // LTBKSLASH_FILL
		[]byte{0x22, 0x22, 0xff, 0x22, 0x22, 0x22, 0xff, 0x22}, // HATCH_FILL
		[]byte{0x81, 0x42, 0x24, 0x18, 0x18, 0x24, 0x42, 0x81}, // XHATCH_FILL
		[]byte{0x11, 0x44, 0x11, 0x44, 0x11, 0x44, 0x11, 0x44}, // INTERLEAVE_FILL
		[]byte{0x10, 0x00, 0x01, 0x00, 0x10, 0x00, 0x01, 0x00}, // WIDE_DOT_FILL
		[]byte{0x11, 0x00, 0x44, 0x00, 0x11, 0x00, 0x44, 0x00}, // CLOSE_DOT_FILL
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // USER_FILL
	}
)

func copyPalette(in color.Palette) (out color.Palette) {
	out = make(color.Palette, len(in))
	copy(out, in)
	return
}

var _ image.Image = (*Canvas)(nil)
