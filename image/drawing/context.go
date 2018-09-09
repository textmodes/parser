package drawing

import (
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype/raster"
)

type LineCap int

const (
	LineCapRound LineCap = iota
	LineCapButt
	LineCapSquare
)

type LineJoin int

const (
	LineJoinRound LineJoin = iota
	LineJoinBevel
)

type FillRule int

const (
	FillRuleWinding FillRule = iota
	FillRuleEvenOdd
)

// Context for drawing
type Context struct {
	width, height int
	//im            *image.RGBA
	im            image.Image
	mask          *image.Alpha
	color         color.Color
	fillPattern   Pattern
	strokePattern Pattern
	fillRule      FillRule
	fillPath      raster.Path
	strokePath    raster.Path
	start         Point
	current       Point
	hasCurrent    bool
	dashes        []float64
	lineWidth     float64
	lineCap       LineCap
	lineJoin      LineJoin
	matrix        Matrix
}

// NewContext returns a new drawing context for the RGBA image.
func NewContext(im image.Image) *Context {
	return &Context{
		width:         im.Bounds().Size().X,
		height:        im.Bounds().Size().Y,
		im:            im,
		color:         color.Transparent,
		fillPattern:   defaultFillStyle,
		strokePattern: defaultStrokeStyle,
		lineWidth:     1,
		matrix:        Identity(),
	}
}

// Image is the resulting image.
func (c *Context) Image() image.Image {
	return c.im
}

// Color sets the drawing color.
func (c *Context) Color(color color.Color) {
	c.color = color
}

// FillStyle sets the fill pattern.
func (c *Context) FillStyle(p Pattern) {
	if s, ok := p.(*solid); ok {
		c.color = s.Color
	}
	c.fillPattern = p
}

// StrokeStyle sets the stroke pattern.
func (c *Context) StrokeStyle(p Pattern) {
	c.strokePattern = p
}

// Matrix sets a new transformation matrix.
func (c *Context) Matrix(matrix Matrix) {
	c.matrix = matrix
}

// Goto moves the drawing cursor.
func (c *Context) Goto(x, y float64) {
	if c.hasCurrent {
		c.fillPath.Add1(c.start.Fixed())
	}
	x, y = c.matrix.Transform(x, y)
	point := Point{x, y}
	c.strokePath.Start(point.Fixed())
	c.fillPath.Start(point.Fixed())
	c.start = point
	c.current = point
	c.hasCurrent = true
}

// LineTo draws a line from the current position to the given point.
func (c *Context) LineTo(x, y float64) {
	if !c.hasCurrent {
		c.Goto(x, y)
	} else {
		x, y = c.matrix.Transform(x, y)
		point := Point{x, y}
		c.strokePath.Add1(point.Fixed())
		c.fillPath.Add1(point.Fixed())
		c.current = point
	}
}

// QuadraticTo adds a quadratic bezier curve to the current path starting at
// the current point. If there is no current point, it first performs
// MoveTo(x1, y1)
func (c *Context) QuadraticTo(x1, y1, x2, y2 float64) {
	if !c.hasCurrent {
		c.Goto(x1, y1)
	}
	x1, y1 = c.matrix.Transform(x1, y1)
	x2, y2 = c.matrix.Transform(x2, y2)
	p1 := Point{x1, y1}
	p2 := Point{x2, y2}
	c.strokePath.Add2(p1.Fixed(), p2.Fixed())
	c.fillPath.Add2(p1.Fixed(), p2.Fixed())
	c.current = p2
}

// CubicTo adds a cubic bezier curve to the current path starting at the
// current point. If there is no current point, it first performs
// MoveTo(x1, y1). Because freetype/raster does not support cubic beziers,
// this is emulated with many small line segments.
func (c *Context) CubicTo(x1, y1, x2, y2, x3, y3 float64) {
	if !c.hasCurrent {
		c.Goto(x1, y1)
	}
	x0, y0 := c.current.X, c.current.Y
	x1, y1 = c.matrix.Transform(x1, y1)
	x2, y2 = c.matrix.Transform(x2, y2)
	x3, y3 = c.matrix.Transform(x3, y3)
	points := CubicBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	previous := c.current.Fixed()
	for _, p := range points[1:] {
		f := p.Fixed()
		if f == previous {
			// TODO: this fixes some rendering issues but not all
			continue
		}
		previous = f
		c.strokePath.Add1(f)
		c.fillPath.Add1(f)
		c.current = p
	}
}

func (c *Context) ClosePath() {
	if !c.hasCurrent {
		return
	}
	c.strokePath.Add1(c.start.Fixed())
	c.fillPath.Add1(c.start.Fixed())
	c.current = c.start
}

func (c *Context) ClearPath() {
	c.strokePath.Clear()
	c.fillPath.Clear()
	c.hasCurrent = false
}

func (c *Context) NewSubPath() {
	if c.hasCurrent {
		c.fillPath.Add1(c.start.Fixed())
	}
	c.hasCurrent = false
}

func (c *Context) DrawEllipse(x, y, rx, ry float64) {
	c.NewSubPath()
	c.DrawEllipticalArc(x, y, rx, ry, 0, 2*math.Pi)
	c.ClosePath()
}

func (c *Context) DrawEllipticalArc(x, y, rx, ry, angle1, angle2 float64) {
	const n = 16
	for i := 0; i < n; i++ {
		p1 := float64(i+0) / n
		p2 := float64(i+1) / n
		a1 := angle1 + (angle2-angle1)*p1
		a2 := angle1 + (angle2-angle1)*p2
		x0 := x + rx*math.Cos(a1)
		y0 := y + ry*math.Sin(a1)
		x1 := x + rx*math.Cos(a1+(a2-a1)/2)
		y1 := y + ry*math.Sin(a1+(a2-a1)/2)
		x2 := x + rx*math.Cos(a2)
		y2 := y + ry*math.Sin(a2)
		cx := 2*x1 - x0/2 - x2/2
		cy := 2*y1 - y0/2 - y2/2
		if i == 0 && !c.hasCurrent {
			c.Goto(x0, y0)
		}
		c.QuadraticTo(cx, cy, x2, y2)
	}
}

// Path Drawing

func (c *Context) capper() raster.Capper {
	switch c.lineCap {
	case LineCapButt:
		return raster.ButtCapper
	case LineCapRound:
		return raster.RoundCapper
	case LineCapSquare:
		return raster.SquareCapper
	}
	return nil
}

func (c *Context) joiner() raster.Joiner {
	switch c.lineJoin {
	case LineJoinBevel:
		return raster.BevelJoiner
	case LineJoinRound:
		return raster.RoundJoiner
	}
	return nil
}

func (c *Context) stroke(painter raster.Painter) {
	path := c.strokePath
	if len(c.dashes) > 0 {
		path = dashed(path, c.dashes)
	} else {
		// TODO: this is a temporary workaround to remove tiny segments
		// that result in rendering issues
		path = rasterPath(flattenPath(path))
	}
	r := raster.NewRasterizer(c.width, c.height)
	r.UseNonZeroWinding = true
	r.AddStroke(path, fix(c.lineWidth), c.capper(), c.joiner())
	r.Rasterize(painter)
}

// StrokePreserve strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is preserved after this
// operation.
func (c *Context) StrokePreserve() {
	painter := newPatternPainter(c.im, c.mask, c.strokePattern)
	c.stroke(painter)
}

// Stroke strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is cleared after this
// operation.
func (c *Context) Stroke() {
	c.StrokePreserve()
	c.ClearPath()
}

func (c *Context) fill(painter raster.Painter) {
	path := c.fillPath
	if c.hasCurrent {
		path = make(raster.Path, len(c.fillPath))
		copy(path, c.fillPath)
		path.Add1(c.start.Fixed())
	}
	r := raster.NewRasterizer(c.width, c.height)
	r.UseNonZeroWinding = c.fillRule == FillRuleWinding
	r.AddPath(path)
	r.Rasterize(painter)
}

// FillPreserve fills the current path with the current color. Open subpaths
// are implicity closed. The path is preserved after this operation.
func (c *Context) FillPreserve() {
	painter := newPatternPainter(c.im, c.mask, c.fillPattern)
	c.fill(painter)
}

// Fill fills the current path with the current color. Open subpaths
// are implicity closed. The path is cleared after this operation.
func (c *Context) Fill() {
	c.FillPreserve()
	c.ClearPath()
}
