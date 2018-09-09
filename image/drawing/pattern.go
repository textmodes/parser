package drawing

import (
	"image"
	"image/color"

	"github.com/golang/freetype/raster"
)

var (
	defaultFillStyle   = NewSolidPattern(color.White)
	defaultStrokeStyle = NewSolidPattern(color.Black)
)

type RepeatOp byte

const (
	RepeatBoth RepeatOp = iota
	RepeatX
	RepeatY
	RepeatNone
)

type Pattern interface {
	ColorAt(x, y int) color.Color
}

type solid struct {
	color.Color
}

func NewSolidPattern(c color.Color) Pattern {
	return solid{c}
}

func (p solid) ColorAt(x, y int) color.Color {
	return color.Color(p)
}

type surface struct {
	im image.Image
	op RepeatOp
}

func (p surface) ColorAt(x, y int) color.Color {
	b := p.im.Bounds()

	switch p.op {
	case RepeatX:
		if y >= b.Dy() {
			return color.Transparent
		}
	case RepeatY:
		if x >= b.Dx() {
			return color.Transparent
		}
	case RepeatNone:
		if x >= b.Dx() || y >= b.Dy() {
			return color.Transparent
		}
	}

	x = x%b.Dx() + b.Min.X
	y = y%b.Dy() + b.Min.Y
	return p.im.At(x, y)
}

func NewSurfacePattern(im image.Image, op RepeatOp) Pattern {
	return &surface{im, op}
}

type patternPainter struct {
	im   image.Image
	mask *image.Alpha
	p    Pattern
}

// Paint satisfies the Painter interface.
func (r *patternPainter) Paint(ss []raster.Span, done bool) {
	b := r.im.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		const m = 1<<16 - 1
		/*
			y := s.Y - r.im.Rect.Min.Y
			x0 := s.X0 - r.im.Rect.Min.X
			// RGBAPainter.Paint() in $GOPATH/src/github.com/golang/freetype/raster/paint.go
			i0 := (s.Y-r.im.Rect.Min.Y)*r.im.Stride + (s.X0-r.im.Rect.Min.X)*4
		*/
		y := s.Y - b.Min.Y
		x0 := s.X0 - b.Min.X
		i0 := (s.Y - b.Min.Y) + (s.X0 - b.Min.X)
		i1 := i0 + (s.X1-s.X0)*4
		for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
			ma := s.Alpha
			if r.mask != nil {
				ma = ma * uint32(r.mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
			}
			c := r.p.ColorAt(x, y)
			cr, cg, cb, ca := c.RGBA()
			d := r.im.At(x, y)
			dr, dg, db, da := d.RGBA()
			a := (m - (ca * ma / m)) * 0x101
			/*
				r.im.Pix[i+0] = uint8((dr*a + cr*ma) / m >> 8)
				r.im.Pix[i+1] = uint8((dg*a + cg*ma) / m >> 8)
				r.im.Pix[i+2] = uint8((db*a + cb*ma) / m >> 8)
				r.im.Pix[i+3] = uint8((da*a + ca*ma) / m >> 8)
			*/
			switch dst := r.im.(type) {
			case *image.RGBA:
				/*
					if da == 0xffff {
						dst.Set(x, y, color.RGBA{
							R: uint8(dr >> 8),
							G: uint8(dg >> 8),
							B: uint8(db >> 8),
							A: uint8(da >> 8),
						})
					} else {
				*/
				dst.Set(x, y, color.RGBA{
					R: uint8((dr*a + cr*ma) / m >> 8),
					G: uint8((dg*a + cg*ma) / m >> 8),
					B: uint8((db*a + cb*ma) / m >> 8),
					A: uint8((da*a + ca*ma) / m >> 8),
				})
				//}
			}
		}
	}
}

func newPatternPainter(im image.Image, mask *image.Alpha, p Pattern) *patternPainter {
	return &patternPainter{im, mask, p}
}

var _ raster.Painter = (*patternPainter)(nil)
