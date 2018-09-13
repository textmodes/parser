package chargen

import (
	"image"
	"image/color"
)

// Filter translates coordinates in a Mask.
type Filter interface {
	// At returns the color at (x, y)
	At(x, y int) color.Color
}

type filterItalics struct {
	Mask
	size   image.Point
	bounds image.Rectangle
}

func (filter filterItalics) At(x, y int) color.Color {
	var (
		dy = (y % filter.size.Y)                 // delta X
		ox = (x / filter.size.X) * filter.size.X // origin X
		sx = filter.size.X >> 2                  // step x
		sy = filter.size.Y >> 2                  // step y
		fx = x                                   // final X
		fy = y                                   // final Y
	)
	if fx = fx + sx - sy + dy/sy; fx < ox {
		return Transparent
	}
	return filter.Mask.At(fx, fy)
}

func (filter filterItalics) Bounds() image.Rectangle {
	return filter.bounds
}

func (filter filterItalics) SubMask(r image.Rectangle) Mask {
	/*
		if filter.Mask == nil {
			return nil
		}
		r = r.Intersect(filter.bounds)
		if r.Empty() {
			return nil
		}
		return filterItalics{
			Mask:   filter.Mask,
			size:   filter.size,
			bounds: r,
		}
	*/
	return Italics(filter.Mask.SubMask(r))
}

// Italics slants the mask to the right.
/*

 01234567      01234567
0________  →  0________
1________  →  1________
2________  →  2________
3________  →  3________
4________  →  4________
5________  →  5________
6________  →  6________
7_####_#_  →  7___####_
8##__##__  →  8_##__##_
9##__##__  →  9_##__##_
a##__##__  →  a_##__##_
b_#####__  →  b__#####_
c____##__  →  c____##__
d##__##__  →  d##__##__
e_####___  →  e_####___
f________  →  f________

 01234567      01234567
0________  →  0________
1________  →  1________
2###_____  →  2___###__
3_##_____  →  3____##__
4_##_____  →  4___##___
5_##_____  →  5___##___
6_##_____  →  6___##___
7_##_##__  →  7___##_##
8_###_##_  →  8__###_##
9_##__##_  →  9__##__##
a_##__##_  →  a__##__##
b###__##_  →  b_###__##
c________  →  c________
d________  →  d________
e________  →  e________
f________  →  f________

*/
func Italics(mask Mask) Mask {
	return filterItalics{
		Mask:   mask,
		size:   mask.CharacterSize(),
		bounds: mask.Bounds(),
	}
}

type filterCharacterRounding struct {
	mask   Mask
	size   image.Point
	bounds image.Rectangle
}

func (filter filterCharacterRounding) At(x, y int) color.Color {
	// Even coordinates?
	x1, y1 := x&1, y&1

	// Scale to match source mask
	x >>= 1
	y >>= 1

	if isOpaque(filter.mask.At(x, y)) {
		// Source pixel is opaque
		return Opaque
	}

	// Source pixel "xy" is transparent; let's check our neighbors:
	/*
		+--+--+--+
		|nw|n |ne|
		+--+--+--+
		|w |xy|e |
		+--+--+--+
		|sw|s |se|
		+--+--+--+
	*/

	// Check north-west
	if x1 == 0 && y1 == 0 && !isOpaque(filter.mask.At(x-1, y-1)) && isOpaque(filter.mask.At(x, y-1)) && isOpaque(filter.mask.At(x-1, y)) {
		return Opaque
	}

	// Check north-east
	if x1 == 0 && y1 == 1 && !isOpaque(filter.mask.At(x-1, y+1)) && isOpaque(filter.mask.At(x, y+1)) && isOpaque(filter.mask.At(x-1, y)) {
		return Opaque
	}

	// Check south-west
	if x1 == 1 && y1 == 0 && !isOpaque(filter.mask.At(x+1, y-1)) && isOpaque(filter.mask.At(x, y-1)) && isOpaque(filter.mask.At(x+1, y)) {
		return Opaque
	}

	// Check south-east
	if x1 == 1 && y1 == 1 && !isOpaque(filter.mask.At(x+1, y+1)) && isOpaque(filter.mask.At(x, y+1)) && isOpaque(filter.mask.At(x+1, y)) {
		return Opaque
	}

	return Transparent
}

func isOpaque(c color.Color) bool {
	if a, ok := c.(*color.Alpha); ok {
		return a.A != 0
	}
	_, _, _, a := c.RGBA()
	return (a >> 8) == 0xff
}

func (filter filterCharacterRounding) Characters() uint16 {
	return filter.mask.Characters()
}

func (filter filterCharacterRounding) CharacterSize() image.Point {
	return filter.size
}

func (filter filterCharacterRounding) ColorModel() color.Model {
	return filter.mask.ColorModel()
}

func (filter filterCharacterRounding) Bounds() image.Rectangle {
	return filter.bounds
}

// SubMask returns a mask representing the portion of the mask visible
// through r. The returned value shares th mask with the original image.
func (filter filterCharacterRounding) SubMask(r image.Rectangle) Mask {
	/*
		if filter.mask == nil {
			return nil
		}
		i := r.Intersect(filter.bounds)
		if i.Empty() {
			return nil
		}
		return filterCharacterRounding{
			mask:   filter.mask,
			size:   filter.size,
			bounds: i,
		}
	*/
	return RoundCharacters(filter.mask.SubMask(image.Rect(
		r.Min.X>>1,
		r.Min.Y>>1,
		r.Max.X>>1,
		r.Max.Y>>1,
	)))
}

// RoundCharacters scales the mask by factor 1 and rounds the characters.
/*

  Font smoothing goes like this, imagine a diagonal line in the original
  like such:

    +--+--+							+--+--+							+--+--+
    |  |##|             |##|  |             |##|  |
    +--+--+             +--+--+             +--+--+
    |##|  |             |  |##|             |##|##|
    +--+--+             +--+--+             +--+--+

  Scaling this up without smoothing, would result in:

    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |  |  |##|##|       |##|##|  |  |       |##|##|  |  |
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |  |  |##|##|       |##|##|  |  |       |##|##|  |  |
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |##|##|  |  |       |  |  |##|##|       |##|##|##|##|
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |##|##|  |  |       |  |  |##|##|       |##|##|##|##|
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+

  With smoothing enabled (notice the last block does *not* smooth):

    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |  |  |##|##|       |##|##|  |  |       |##|##|  |  |
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |  |##|##|##|       |##|##|##|  |       |##|##|  |  |
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |##|##|##|  |       |  |##|##|##|       |##|##|##|##|
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+
    |##|##|  |  |       |  |  |##|##|       |##|##|##|##|
    +--+--+--+--+       +--+--+--+--+       +--+--+--+--+

  Half dot is inserted before or after a whole dot in the presence of
  a diagonal in the matrix.

*/
func RoundCharacters(mask Mask) Mask {
	size := mask.CharacterSize()
	size.X <<= 1
	size.Y <<= 1
	bounds := mask.Bounds()
	bounds.Min.X <<= 1
	bounds.Min.Y <<= 1
	bounds.Max.X <<= 1
	bounds.Max.Y <<= 1
	return filterCharacterRounding{
		mask:   mask,
		size:   size,
		bounds: bounds,
	}
}

type columnAdder struct {
	Mask
	size   image.Point
	bounds image.Rectangle
}

func (filter columnAdder) At(x, y int) color.Color {
	if x%filter.size.X == filter.size.X-1 {
		return Transparent
	}
	x = (x * filter.size.X) / (filter.size.X - 1)
	return filter.Mask.At(x, y)
}

func (filter columnAdder) Bounds() image.Rectangle {
	return filter.bounds
}

func (filter columnAdder) CharacterSize() image.Point {
	return filter.size
}

// AddColumn adds a blank column to the right hand side of each character.
func AddColumn(mask Mask) Mask {
	var (
		size   = mask.CharacterSize()
		chars  = mask.Characters()
		bounds = mask.Bounds()
	)
	size.X++
	bounds.Max.X += int(chars)
	return &columnAdder{
		Mask:   mask,
		size:   size,
		bounds: bounds,
	}
}
